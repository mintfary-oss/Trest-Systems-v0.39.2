package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mintfary-oss/trest-sistems/internal/auth"
)

type createObjectRequest struct {
	ProjectID, Name, Address string
	AreaM2                   float64
}
type createEstimateItemRequest struct {
	ProjectID, ObjectID, Name, Category, Unit string
	Quantity, UnitPrice                       float64
}
type createOrderRequest struct {
	ProjectID, ObjectID, Title string
	Amount                     float64
}
type createOrderItemRequest struct {
	Name, Unit          string
	Quantity, UnitPrice float64
}
type createBidRequest struct {
	BidType string
	Amount  float64
	Comment string
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return false
	}
	return true
}
func userClaims(r *http.Request) (*auth.Claims, bool) {
	c, ok := claimsFromContext(r.Context())
	if !ok {
		return nil, false
	}
	return &c, true
}

func (s *Server) objects(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id, project_id, name, address, area_m2, created_at FROM objects WHERE project_id=$1 ORDER BY created_at DESC`, projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, p, n, a string
		var area, created any
		if err := rows.Scan(&id, &p, &n, &a, &area, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "project_id": p, "name": n, "address": a, "area_m2": area, "created_at": created})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) createObject(w http.ResponseWriter, r *http.Request) {
	c, ok := userClaims(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in createObjectRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.Name == "" || in.ProjectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and name are required"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO objects(project_id,name,address,area_m2) SELECT $1,$2,$3,$4 WHERE EXISTS(SELECT 1 FROM projects WHERE id=$1 AND owner_id=$5) RETURNING id`, in.ProjectID, in.Name, in.Address, in.AreaM2, c.UserID).Scan(&id)
	if err != nil {
		writeJSON(w, 403, map[string]string{"error": "project not found or not owned by user"})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "project_id": in.ProjectID, "name": in.Name, "address": in.Address, "area_m2": in.AreaM2})
}

func (s *Server) estimate(w http.ResponseWriter, r *http.Request) {
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	if projectID == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,object_id,name,category,unit,quantity,unit_price,approved,created_at FROM estimate_items WHERE project_id=$1 ORDER BY created_at`, projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id string
		var oid any
		var name, cat, unit string
		var q, p any
		var appr bool
		var created any
		if err := rows.Scan(&id, &oid, &name, &cat, &unit, &q, &p, &appr, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "object_id": oid, "name": name, "category": cat, "unit": unit, "quantity": q, "unit_price": p, "approved": appr, "created_at": created})
	}
	if err := rows.Err(); err != nil {
		writeError(w, err)
		return
	}
	var total any
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COALESCE(SUM(quantity*unit_price),0) FROM estimate_items WHERE project_id=$1`, projectID).Scan(&total); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"project_id": projectID, "items": out, "total": total})
}

func (s *Server) createEstimateItem(w http.ResponseWriter, r *http.Request) {
	c, ok := userClaims(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in createEstimateItemRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.ProjectID == "" || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and name are required"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO estimate_items(project_id,object_id,name,category,unit,quantity,unit_price) SELECT $1,$2,$3,$4,$5,$6,$7 WHERE EXISTS(SELECT 1 FROM projects WHERE id=$1 AND owner_id=$8) RETURNING id`, in.ProjectID, nilIfEmpty(in.ObjectID), in.Name, defaultStr(in.Category, "general"), defaultStr(in.Unit, "pcs"), in.Quantity, in.UnitPrice, c.UserID).Scan(&id)
	if err != nil {
		writeJSON(w, 403, map[string]string{"error": "project not found or not owned by user"})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}
func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
func defaultStr(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func (s *Server) createLegacyMarketplaceOrder(w http.ResponseWriter, r *http.Request) {
	c, ok := userClaims(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in createOrderRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	if in.ProjectID == "" || in.Title == "" {
		writeJSON(w, 400, map[string]string{"error": "project_id and title are required"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO orders(project_id,object_id,title,amount) SELECT $1,$2,$3,$4 WHERE EXISTS(SELECT 1 FROM projects WHERE id=$1 AND owner_id=$5) RETURNING id`, in.ProjectID, nilIfEmpty(in.ObjectID), in.Title, in.Amount, c.UserID).Scan(&id)
	if err != nil {
		writeJSON(w, 403, map[string]string{"error": "project not found or not owned by user"})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "draft"})
}

func (s *Server) orderItems(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" {
		writeJSON(w, 400, map[string]string{"error": "order_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT id,name,unit,quantity,unit_price,created_at FROM order_items WHERE order_id=$1 ORDER BY created_at`, orderID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, unit string
		var q, p, created any
		if err := rows.Scan(&id, &name, &unit, &q, &p, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "unit": unit, "quantity": q, "unit_price": p, "created_at": created})
	}
	writeJSON(w, 200, out)
}
func (s *Server) createOrderItem(w http.ResponseWriter, r *http.Request) {
	c, ok := userClaims(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in createOrderItemRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" || in.Name == "" {
		writeJSON(w, 400, map[string]string{"error": "order_id and name are required"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO order_items(order_id,name,unit,quantity,unit_price) SELECT $1,$2,$3,$4,$5 WHERE EXISTS(SELECT 1 FROM orders o JOIN projects p ON p.id=o.project_id WHERE o.id=$1 AND p.owner_id=$6) RETURNING id`, orderID, in.Name, defaultStr(in.Unit, "pcs"), in.Quantity, in.UnitPrice, c.UserID).Scan(&id)
	if err != nil {
		writeJSON(w, 403, map[string]string{"error": "order not found or not owned by user"})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id})
}

func (s *Server) bids(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" {
		writeJSON(w, 400, map[string]string{"error": "order_id is required"})
		return
	}
	rows, err := s.DB.QueryContext(r.Context(), `SELECT b.id,b.bidder_id,b.bid_type,b.amount,b.comment,b.status,b.created_at,u.name FROM marketplace_bids b JOIN users u ON u.id=b.bidder_id WHERE b.order_id=$1 ORDER BY b.created_at DESC`, orderID)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := make([]map[string]any, 0)
	for rows.Next() {
		var id, bidder, typ, comment, status, name string
		var amount, created any
		if err := rows.Scan(&id, &bidder, &typ, &amount, &comment, &status, &created, &name); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "bidder_id": bidder, "bidder_name": name, "bid_type": typ, "amount": amount, "comment": comment, "status": status, "created_at": created})
	}
	writeJSON(w, 200, out)
}
func (s *Server) createBid(w http.ResponseWriter, r *http.Request) {
	c, ok := userClaims(r)
	if !ok {
		writeJSON(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	var in createBidRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
	if orderID == "" || in.Amount < 0 || (in.BidType != "contractor" && in.BidType != "supplier") {
		writeJSON(w, 400, map[string]string{"error": "valid order_id, bid_type and amount are required"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO marketplace_bids(order_id,bidder_id,bid_type,amount,comment) SELECT $1,$2,$3,$4,$5 WHERE EXISTS(SELECT 1 FROM orders WHERE id=$1 AND status IN ('published','matching')) RETURNING id`, orderID, c.UserID, in.BidType, in.Amount, in.Comment).Scan(&id)
	if err != nil {
		writeJSON(w, 409, map[string]string{"error": "order unavailable or duplicate bid"})
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "submitted"})
}
