package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

type MarketplaceHandler struct{ pool *pgxpool.Pool }

func NewMarketplaceHandler(pool *pgxpool.Pool) *MarketplaceHandler { return &MarketplaceHandler{pool: pool} }

func (h *MarketplaceHandler) ListProducts(ctx context.Context, req *connect.Request[alfqv1.ListProductsRequest]) (*connect.Response[alfqv1.ProductList], error) {
	r := req.Msg
	rows, _ := h.pool.Query(ctx, `
		SELECT id, author_id, author_name, name, category, description, symbol, purchase_type, price, trial_days, rating, purchase_count, created_at
		FROM alfq_products WHERE ($1='' OR category=$1) ORDER BY rating DESC LIMIT $2
	`, r.Category, minInt(r.PageSize, 20))
	defer rows.Close()
	var prods []*alfqv1.Product
	for rows.Next() {
		p := &alfqv1.Product{}
		var ts time.Time
		rows.Scan(&p.Id, &p.AuthorId, &p.AuthorName, &p.Name, &p.Category, &p.Description, &p.Symbol, &p.PurchaseType, &p.Price, &p.TrialDays, &p.Rating, &p.PurchaseCount, &ts)
		p.CreatedAt = ts.Unix()
		prods = append(prods, p)
	}
	return connect.NewResponse(&alfqv1.ProductList{Products: prods}), nil
}

func (h *MarketplaceHandler) PublishProduct(ctx context.Context, req *connect.Request[alfqv1.PublishProductRequest]) (*connect.Response[alfqv1.Product], error) {
	uid := userIDFromHTTP(ctx, req)
	var name string
	h.pool.QueryRow(ctx, "SELECT COALESCE(display_name, email) FROM users WHERE id=$1", uid).Scan(&name)
	r := req.Msg
	p := &alfqv1.Product{AuthorId: uid, AuthorName: name, Name: r.Name, Category: r.Category, Description: r.Description, Symbol: r.Symbol, PurchaseType: r.PurchaseType, Price: r.Price, TrialDays: r.TrialDays}
	var ts time.Time
	h.pool.QueryRow(ctx, `
		INSERT INTO alfq_products (author_id, author_name, name, category, description, symbol, purchase_type, price, trial_days)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at
	`, uid, name, r.Name, r.Category, r.Description, r.Symbol, r.PurchaseType, r.Price, r.TrialDays).Scan(&p.Id, &ts)
	p.CreatedAt = ts.Unix()
	return connect.NewResponse(p), nil
}

func (h *MarketplaceHandler) PurchaseProduct(ctx context.Context, req *connect.Request[alfqv1.PurchaseProductRequest]) (*connect.Response[alfqv1.Purchase], error) {
	uid := userIDFromHTTP(ctx, req)
	var pname string
	h.pool.QueryRow(ctx, "SELECT name FROM alfq_products WHERE id=$1", req.Msg.ProductId).Scan(&pname)
	pur := &alfqv1.Purchase{ProductId: req.Msg.ProductId, ProductName: pname, IsTrial: req.Msg.StartTrial}
	var ts time.Time
	h.pool.QueryRow(ctx, `
		INSERT INTO alfq_purchases (user_id, product_id, is_trial)
		VALUES ($1,$2,$3) RETURNING id, expires_at, created_at
	`, uid, req.Msg.ProductId, req.Msg.StartTrial).Scan(&pur.Id, &pur.ExpiresAt, &ts)
	pur.CreatedAt = ts.Unix()
	h.pool.Exec(ctx, "UPDATE alfq_products SET purchase_count = purchase_count + 1 WHERE id=$1", req.Msg.ProductId)
	return connect.NewResponse(pur), nil
}

func (h *MarketplaceHandler) GetMyProducts(ctx context.Context, req *connect.Request[alfqv1.GetMyProductsRequest]) (*connect.Response[alfqv1.ProductList], error) {
	uid := userIDFromHTTP(ctx, req)
	rows, _ := h.pool.Query(ctx, `
		SELECT id, author_id, author_name, name, category, description, symbol, purchase_type, price, trial_days, rating, purchase_count, created_at
		FROM alfq_products WHERE author_id=$1 ORDER BY created_at DESC
	`, uid)
	defer rows.Close()
	var prods []*alfqv1.Product
	for rows.Next() {
		p := &alfqv1.Product{}
		var ts time.Time
		rows.Scan(&p.Id, &p.AuthorId, &p.AuthorName, &p.Name, &p.Category, &p.Description, &p.Symbol, &p.PurchaseType, &p.Price, &p.TrialDays, &p.Rating, &p.PurchaseCount, &ts)
		p.CreatedAt = ts.Unix()
		prods = append(prods, p)
	}
	return connect.NewResponse(&alfqv1.ProductList{Products: prods}), nil
}

func (h *MarketplaceHandler) GetMyPurchases(ctx context.Context, req *connect.Request[alfqv1.GetMyPurchasesRequest]) (*connect.Response[alfqv1.PurchaseList], error) {
	uid := userIDFromHTTP(ctx, req)
	rows, _ := h.pool.Query(ctx, `
		SELECT id, product_id, (SELECT name FROM alfq_products WHERE id=p.product_id), is_trial, expires_at, created_at
		FROM alfq_purchases p WHERE user_id=$1 ORDER BY created_at DESC
	`, uid)
	defer rows.Close()
	var purs []*alfqv1.Purchase
	for rows.Next() {
		p := &alfqv1.Purchase{}
		var ts time.Time
		rows.Scan(&p.Id, &p.ProductId, &p.ProductName, &p.IsTrial, &p.ExpiresAt, &ts)
		p.CreatedAt = ts.Unix()
		purs = append(purs, p)
	}
	return connect.NewResponse(&alfqv1.PurchaseList{Purchases: purs}), nil
}

var _ antclawv1connect.MarketplaceServiceHandler = (*MarketplaceHandler)(nil)
