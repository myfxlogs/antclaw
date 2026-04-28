package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

// getRedisAddr mirrors worker configuration for API server.
func getRedisAddr() string {
	host := os.Getenv("ANTCLAW_REDIS_HOST")
	if host == "" {
		host = "redis"
	}
	port := os.Getenv("ANTCLAW_REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return fmt.Sprintf("%s:%s", host, port)
}

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
