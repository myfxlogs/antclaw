package rpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	devicev1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/auth"
)

type DeviceHandler struct{ pg *pgxpool.Pool }

func NewDeviceHandler(pg *pgxpool.Pool) *DeviceHandler { return &DeviceHandler{pg: pg} }

func (h *DeviceHandler) ReportDeviceInfo(ctx context.Context, req *connect.Request[devicev1.ReportDeviceInfoRequest]) (*connect.Response[devicev1.ReportDeviceInfoResponse], error) {
	di := req.Msg.GetDeviceInfo()
	if di == nil || di.GetDeviceId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}

	userID, ok := auth.UserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication required"))
	}
	role, _ := auth.RoleFromContext(ctx)

	// 检查设备绑定归属：不允许抢占其他用户的设备
	var ownerID string
	err := h.pg.QueryRow(ctx,
		`SELECT COALESCE(user_id::text,'') FROM devices WHERE device_id=$1`,
		di.GetDeviceId(),
	).Scan(&ownerID)
	if err == nil && ownerID != "" && ownerID != userID {
		log.Printf("ReportDeviceInfo: device=%s already bound to user=%s, rejected for user=%s role=%s",
			di.GetDeviceId(), ownerID, userID, role)
		return nil, connect.NewError(connect.CodePermissionDenied,
			fmt.Errorf("device already bound to another user"))
	}

	_, err = h.pg.Exec(ctx, `
		INSERT INTO devices (device_id, model, brand, os_version, os_type, app_version, build_number,
		                     screen_width, screen_height, network_type, timezone, locale, manufacturer, fingerprint, user_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (device_id) DO UPDATE SET
			model=$2, brand=$3, os_version=$4, os_type=$5, app_version=$6, build_number=$7,
			screen_width=$8, screen_height=$9, network_type=$10, timezone=$11, locale=$12,
			manufacturer=$13, fingerprint=$14, user_id=$15, updated_at=NOW()`,
		di.GetDeviceId(), di.GetModel(), di.GetBrand(), di.GetOsVersion(), di.GetOsType(),
		di.GetAppVersion(), di.GetBuildNumber(), di.GetScreenWidth(), di.GetScreenHeight(),
		di.GetNetworkType(), di.GetTimezone(), di.GetLocale(), di.GetManufacturer(),
		di.GetFingerprint(), userID,
	)
	if err != nil {
		log.Printf("ReportDeviceInfo: db error device=%s user=%s err=%v", di.GetDeviceId(), userID, err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	log.Printf("ReportDeviceInfo: success device=%s user=%s os=%s app=%s",
		di.GetDeviceId(), userID, di.GetOsType(), di.GetAppVersion())
	return connect.NewResponse(&devicev1.ReportDeviceInfoResponse{Success: true}), nil
}

func (h *DeviceHandler) DeleteDeviceInfo(ctx context.Context, req *connect.Request[devicev1.DeleteDeviceInfoRequest]) (*connect.Response[devicev1.DeleteDeviceInfoResponse], error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if req.Msg.GetDeviceId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device_id is required"))
	}
	_, err := h.pg.Exec(ctx, `DELETE FROM devices WHERE device_id = $1`, req.Msg.GetDeviceId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&devicev1.DeleteDeviceInfoResponse{Success: true}), nil
}

func (h *DeviceHandler) ListDevices(ctx context.Context, req *connect.Request[devicev1.ListDevicesRequest]) (*connect.Response[devicev1.ListDevicesResponse], error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	filter := req.Msg.GetOsTypeFilter()
	limit := clampPage(req.Msg.GetPageSize()) + 1

	rows, err := h.pg.Query(ctx, `
		SELECT d.device_id, d.model, d.brand, d.os_version, d.os_type, d.app_version, d.build_number,
		       d.screen_width, d.screen_height, d.network_type, d.timezone, d.locale, d.manufacturer,
		       d.fingerprint, COALESCE(d.user_id::text,''), d.created_at, d.updated_at,
		       COALESCE(u.display_name,''), COALESCE(u.username,''), COALESCE(u.code_id,'')
		  FROM devices d LEFT JOIN users u ON u.id = d.user_id
		 WHERE ($1='' OR d.os_type=$1) ORDER BY d.updated_at DESC LIMIT $2`,
		filter, limit,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()

	var out []*devicev1.DeviceInfo
	for rows.Next() {
		var d devicev1.DeviceInfo
		var uid string
		var ca, ua time.Time
		if rows.Scan(&d.DeviceId, &d.Model, &d.Brand, &d.OsVersion, &d.OsType,
			&d.AppVersion, &d.BuildNumber, &d.ScreenWidth, &d.ScreenHeight,
			&d.NetworkType, &d.Timezone, &d.Locale, &d.Manufacturer,
			&d.Fingerprint, &uid, &ca, &ua, &d.DisplayName, &d.Username, &d.CodeId) == nil {
			d.UserId = uid
			d.CreatedAt = ca.Unix()
			d.UpdatedAt = ua.Unix()
			out = append(out, &d)
		}
	}

	var total int32
	// total 尊重 os_type_filter
	_ = h.pg.QueryRow(ctx, `SELECT COUNT(*) FROM devices WHERE ($1='' OR os_type=$1)`, filter).Scan(&total)
	return connect.NewResponse(&devicev1.ListDevicesResponse{Devices: out, Total: total}), nil
}

// requireAdmin 检查当前 context 中用户是否为管理员，非管理员返回 PermissionDenied。
func requireAdmin(ctx context.Context) error {
	role, ok := auth.RoleFromContext(ctx)
	if !ok || (role != "admin" && role != "super_admin") {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("admin access required"))
	}
	return nil
}

func clampPage(n int32) int32 {
	if n <= 0 || n > 100 {
		return 50
	}
	return n
}

var _ antclawv1connect.DeviceServiceHandler = (*DeviceHandler)(nil)
