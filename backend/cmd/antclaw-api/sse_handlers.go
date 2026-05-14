package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	redisv9 "github.com/redis/go-redis/v9"

	"github.com/antclaw/antclaw/internal/service/presence"
)

// streamSSE 封装 Redis Streams 到 SSE 的通用推送逻辑：
//   - 连接后立即发送一条推送注释路由头，避免代理/curl --max-time 错误超时
//   - 每 15s 发送一个 SSE 维连注释（`: ping`）
//   - XRead 使用 5s block，到期则走心跳分支避免一直阻塞
func streamSSE(w http.ResponseWriter, r *http.Request, rdb *redisv9.Client, streamKey string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, ": connected stream=%s ts=%d\n\n", streamKey, time.Now().Unix())
	flusher.Flush()

	ctx := r.Context()
	lastID := "$"
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix())
			flusher.Flush()
		default:
		}
		streams, err := rdb.XRead(ctx, &redisv9.XReadArgs{
			Streams: []string{streamKey, lastID},
			Block:   5 * time.Second,
			Count:   10,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if err == redisv9.Nil || err.Error() == "redis: nil" {
				continue
			}
			log.Printf("sse %s XRead error: %v", streamKey, err)
			time.Sleep(time.Second)
			continue
		}
		for _, s := range streams {
			for _, msg := range s.Messages {
				lastID = msg.ID
				data, ok := msg.Values["data"].(string)
				if !ok {
					b, _ := json.Marshal(msg.Values)
					data = string(b)
				}
				fmt.Fprintf(w, "id: %s\n", msg.ID)
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}

// jobsEventsHandler streams job events from Redis Streams to the browser via SSE.
func jobsEventsHandler(rdb *redisv9.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamSSE(w, r, rdb, "stream:jobs_events")
	}
}

// auditEventsHandler streams audit log events via SSE.
func auditEventsHandler(rdb *redisv9.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamSSE(w, r, rdb, "stream:audit_events")
	}
}

// alertsEventsHandler 通用告警 SSE：把任意 Redis Stream 转推给浏览器。
// 当上游尚未发布告警时，连接保持打开，前端显示 "等待事件..."；这是预期行为，避免 404。
func alertsEventsHandler(rdb *redisv9.Client, stream string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamSSE(w, r, rdb, stream)
	}
}

// userNotificationsSSE 个人通知 SSE：用 Redis Pub/Sub 实时推送 user:{userID}:notifications。
//
// 鉴权：从 Authorization: Bearer 或 Cookie antclaw_at= 提取 access_token，校验后取 sub 作为 userID。
// 失败 → 401，避免泄露推送频道。
func userNotificationsSSE(rdb *redisv9.Client, pt *presence.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, role, err := extractUserIDFromRequest(r)
		if err != nil {
			log.Printf("SSE notifications: auth failed remote=%s", r.RemoteAddr)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.Printf("SSE notifications: connected user=%s remote=%s role=%s", userID, r.RemoteAddr, role)
		// 管理端用户不纳入在线统计
		if role != "admin" && role != "super_admin" {
			pt.Register(userID, r.RemoteAddr)
			defer func() {
				pt.Unregister(userID)
				log.Printf("SSE notifications: disconnected user=%s", userID)
			}()
		} else {
			defer log.Printf("SSE notifications: admin disconnected user=%s", userID)
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, ": connected user=%s ts=%d\n\n", userID, time.Now().Unix())
		flusher.Flush()

		channel := "user:" + userID + ":notifications"
		ctx := r.Context()
		pubsub := rdb.Subscribe(ctx, channel)
		defer pubsub.Close()
		ch := pubsub.Channel()
		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-heartbeat.C:
				fmt.Fprintf(w, ": ping %d\n\n", time.Now().Unix())
				flusher.Flush()
			case m, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: notification\ndata: %s\n\n", m.Payload)
				flusher.Flush()
			}
		}
	}
}