package rpc

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"connectrpc.com/connect"
	cryptov1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
	redispkg "github.com/antclaw/antclaw/internal/infra/redis"
)

// CryptoConnectHandler 暴露 RSA 公钥等只读加密元数据。
type CryptoConnectHandler struct {
	rsa *cryptopkg.RSAManager
	rdb *redispkg.Client
}

// NewCryptoConnectHandler 创建 handler。
func NewCryptoConnectHandler(rsa *cryptopkg.RSAManager, rdb *redispkg.Client) *CryptoConnectHandler {
	return &CryptoConnectHandler{rsa: rsa, rdb: rdb}
}

func (h *CryptoConnectHandler) GetCryptoPublicKey(ctx context.Context, _ *connect.Request[cryptov1.GetCryptoPublicKeyRequest]) (*connect.Response[cryptov1.GetCryptoPublicKeyResponse], error) {
	_ = ctx
	return connect.NewResponse(&cryptov1.GetCryptoPublicKeyResponse{
		Pem: string(h.rsa.PublicPEM()),
	}), nil
}

// PostEnvelope 解密并验证一个加密信封，并回显明文（base64）作为占位实现。
func (h *CryptoConnectHandler) PostEnvelope(ctx context.Context, req *connect.Request[cryptov1.PostEnvelopeRequest]) (*connect.Response[cryptov1.PostEnvelopeResponse], error) {
	bodyB64 := req.Msg.GetBodyB64()
	ts := req.Msg.GetTs()
	nonce := req.Msg.GetNonce()
	sig := req.Msg.GetSig()

	// 从 Authorization 取 token
	token := ""
	if auth := req.Header().Get("Authorization"); len(auth) > 7 {
		token = auth[7:]
	}
	// 先对外层进行签名验证（签名的数据是 bodyB64 原文）
	if err := cryptopkg.VerifyRequestSignature(ctx, h.rdb.Raw(), token, ts, nonce, []byte(bodyB64), sig); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// 解析并解密 envelope
	raw, err := base64.StdEncoding.DecodeString(bodyB64)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	var env cryptopkg.HybridEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	plain, err := h.rsa.Decrypt(env)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&cryptov1.PostEnvelopeResponse{BodyB64: base64.StdEncoding.EncodeToString(plain)}), nil
}

var _ antclawv1connect.CryptoServiceHandler = (*CryptoConnectHandler)(nil)
