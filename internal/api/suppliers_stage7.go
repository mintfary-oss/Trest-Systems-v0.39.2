package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) createSupplierApplication(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct{ OrganizationID, Note string }
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.OrganizationID) == "" {
		writeJSON(w, 400, map[string]string{"error": "organization_id is required"})
		return
	}
	var typ string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT type FROM organizations WHERE id=$1`, in.OrganizationID).Scan(&typ); err != nil {
		writeJSON(w, 404, map[string]string{"error": "organization not found"})
		return
	}
	if typ != "supplier" && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "organization must be supplier type"})
		return
	}
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO supplier_applications(organization_id,applicant_user_id,note) VALUES($1,$2,$3) ON CONFLICT(organization_id,applicant_user_id) DO UPDATE SET status='pending',note=EXCLUDED.note,reviewed_at=NULL RETURNING id`, in.OrganizationID, c.UserID, in.Note).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "pending"})
}

func (s *Server) upsertSupplierProfile(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct {
		OrganizationID, Summary     string
		Categories, DeliveryRegions []string
		DeliveryTerms               map[string]any
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.OrganizationID) == "" {
		writeJSON(w, 400, map[string]string{"error": "organization_id is required"})
		return
	}
	var allowed int
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM organization_members WHERE organization_id=$1 AND user_id=$2 AND membership_role IN ('owner','admin')`, in.OrganizationID, c.UserID).Scan(&allowed); err != nil || (allowed == 0 && c.Role != "admin") {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	cats, _ := json.Marshal(in.Categories)
	regions, _ := json.Marshal(in.DeliveryRegions)
	terms, _ := json.Marshal(in.DeliveryTerms)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO supplier_profiles(organization_id,summary,categories,delivery_regions,delivery_terms) VALUES($1,$2,$3,$4,$5) ON CONFLICT(organization_id) DO UPDATE SET summary=EXCLUDED.summary,categories=EXCLUDED.categories,delivery_regions=EXCLUDED.delivery_regions,delivery_terms=EXCLUDED.delivery_terms,updated_at=now() RETURNING id`, in.OrganizationID, in.Summary, cats, regions, terms).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "organization_id": in.OrganizationID, "verification_status": "pending"})
}

func (s *Server) verifySupplier(w http.ResponseWriter, r *http.Request) {
	var in struct{ SupplierID, Status string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))
	if in.Status != "verified" && in.Status != "rejected" && in.Status != "suspended" {
		writeJSON(w, 400, map[string]string{"error": "invalid status"})
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `UPDATE supplier_profiles SET verification_status=$1,active=CASE WHEN $1='suspended' THEN false WHEN $1='verified' THEN true ELSE active END,updated_at=now() WHERE id=$2`, in.Status, in.SupplierID)
	if err != nil {
		writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, map[string]string{"error": "supplier not found"})
		return
	}
	writeJSON(w, 200, map[string]string{"status": in.Status})
}

func (s *Server) suppliers(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	category := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("category")))
	q := `SELECT sp.id,sp.organization_id,o.legal_name,sp.summary,sp.categories,sp.delivery_regions,sp.delivery_terms,sp.verification_status,sp.active,sp.created_at FROM supplier_profiles sp JOIN organizations o ON o.id=sp.organization_id WHERE 1=1`
	args := []any{}
	n := 1
	if status != "" {
		q += ` AND sp.verification_status=$` + strconv.Itoa(n)
		args = append(args, status)
		n++
	}
	if category != "" {
		q += ` AND sp.categories @> $` + strconv.Itoa(n) + `::jsonb`
		args = append(args, `["`+category+`"]`)
		n++
	}
	q += ` ORDER BY sp.created_at DESC LIMIT 100`
	rows, err := s.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, org, name, summary, status string
		var cats, regions, terms any
		var active bool
		var created any
		if err := rows.Scan(&id, &org, &name, &summary, &cats, &regions, &terms, &status, &active, &created); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "organization_id": org, "legal_name": name, "summary": summary, "categories": cats, "delivery_regions": regions, "delivery_terms": terms, "verification_status": status, "active": active, "created_at": created})
	}
	writeJSON(w, 200, out)
}

func (s *Server) createSupplierOffer(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	var in struct {
		SupplierID, SKU, Name, Category, Unit, Currency string
		Price, MinOrderQuantity, StockQuantity          float64
		LeadTimeDays                                    int
		Metadata                                        map[string]any
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid json"})
		return
	}
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	if in.SupplierID == "" || in.SKU == "" || in.Name == "" || in.Price < 0 || in.MinOrderQuantity <= 0 || in.StockQuantity < 0 || in.LeadTimeDays < 0 {
		writeJSON(w, 400, map[string]string{"error": "invalid supplier offer"})
		return
	}
	var allowed int
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM supplier_profiles sp JOIN organization_members om ON om.organization_id=sp.organization_id WHERE sp.id=$1 AND om.user_id=$2 AND om.membership_role IN ('owner','admin')`, in.SupplierID, c.UserID).Scan(&allowed); err != nil || (allowed == 0 && c.Role != "admin") {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	if in.Currency == "" {
		in.Currency = "RUB"
	}
	if in.Unit == "" {
		in.Unit = "pcs"
	}
	if in.Category == "" {
		in.Category = "general"
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}
	meta, _ := json.Marshal(in.Metadata)
	var id string
	err := s.DB.QueryRowContext(r.Context(), `INSERT INTO supplier_offers(supplier_id,sku,name,category,unit,price,currency,min_order_quantity,stock_quantity,lead_time_days,metadata) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(supplier_id,sku) DO UPDATE SET name=EXCLUDED.name,category=EXCLUDED.category,unit=EXCLUDED.unit,price=EXCLUDED.price,currency=EXCLUDED.currency,min_order_quantity=EXCLUDED.min_order_quantity,stock_quantity=EXCLUDED.stock_quantity,lead_time_days=EXCLUDED.lead_time_days,metadata=EXCLUDED.metadata,updated_at=now() RETURNING id`, in.SupplierID, in.SKU, in.Name, in.Category, in.Unit, in.Price, in.Currency, in.MinOrderQuantity, in.StockQuantity, in.LeadTimeDays, meta).Scan(&id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "status": "draft"})
}

func (s *Server) publishSupplierOffer(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	id := strings.TrimSpace(r.PathValue("offerID"))
	var allowed int
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM supplier_offers so JOIN supplier_profiles sp ON sp.id=so.supplier_id JOIN organization_members om ON om.organization_id=sp.organization_id WHERE so.id=$1 AND om.user_id=$2 AND om.membership_role IN ('owner','admin')`, id, c.UserID).Scan(&allowed); err != nil || (allowed == 0 && c.Role != "admin") {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	res, err := s.DB.ExecContext(r.Context(), `UPDATE supplier_offers SET status='published',updated_at=now() WHERE id=$1`, id)
	if err != nil {
		writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeJSON(w, 404, map[string]string{"error": "offer not found"})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "published"})
}

func (s *Server) supplierOffers(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("category")))
	region := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("region")))
	limit := 100
	q := `SELECT so.id,so.supplier_id,o.legal_name,so.sku,so.name,so.category,so.unit,so.price,so.currency,so.min_order_quantity,so.stock_quantity,so.lead_time_days,so.status FROM supplier_offers so JOIN supplier_profiles sp ON sp.id=so.supplier_id JOIN organizations o ON o.id=sp.organization_id WHERE so.status='published' AND sp.verification_status='verified' AND sp.active=true`
	args := []any{}
	n := 1
	if category != "" {
		q += ` AND so.category=$` + strconv.Itoa(n)
		args = append(args, category)
		n++
	}
	if region != "" {
		q += ` AND sp.delivery_regions @> $` + strconv.Itoa(n) + `::jsonb`
		args = append(args, `["`+region+`"]`)
		n++
	}
	q += ` ORDER BY so.price ASC LIMIT ` + strconv.Itoa(limit)
	rows, err := s.DB.QueryContext(r.Context(), q, args...)
	if err != nil {
		writeError(w, err)
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, sid, legal, sku, name, cat, unit, currency, status string
		var price, minq, stock any
		var lead int
		if err := rows.Scan(&id, &sid, &legal, &sku, &name, &cat, &unit, &price, &currency, &minq, &stock, &lead, &status); err != nil {
			writeError(w, err)
			return
		}
		out = append(out, map[string]any{"id": id, "supplier_id": sid, "legal_name": legal, "sku": sku, "name": name, "category": cat, "unit": unit, "price": price, "currency": currency, "min_order_quantity": minq, "stock_quantity": stock, "lead_time_days": lead, "status": status})
	}
	writeJSON(w, 200, out)
}

func (s *Server) attachSupplierOfferToOrder(w http.ResponseWriter, r *http.Request) {
	c, _ := claimsFromContext(r.Context())
	orderID := strings.TrimSpace(r.PathValue("id"))
	var in struct {
		OfferID        string  `json:"offer_id"`
		Quantity       float64 `json:"quantity"`
		IdempotencyKey string  `json:"idempotency_key"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil || in.OfferID == "" || in.Quantity <= 0 {
		writeJSON(w, 400, map[string]string{"error": "offer_id and positive quantity are required"})
		return
	}
	var owner string
	if err := s.DB.QueryRowContext(r.Context(), `SELECT COALESCE(e.created_by,p.owner_id) FROM orders o JOIN projects p ON p.id=o.project_id LEFT JOIN estimate_versions ev ON ev.id=o.estimate_version_id LEFT JOIN estimates e ON e.id=ev.estimate_id WHERE o.id=$1`, orderID).Scan(&owner); err != nil {
		writeJSON(w, 404, map[string]string{"error": "order not found"})
		return
	}
	if owner != c.UserID && c.Role != "admin" {
		writeJSON(w, 403, map[string]string{"error": "forbidden"})
		return
	}
	var supplierID string
	var price, stock, minq float64
	var status, verification string
	var active bool
	err := s.DB.QueryRowContext(r.Context(), `SELECT so.supplier_id,so.price,so.stock_quantity,so.min_order_quantity,so.status,sp.verification_status,sp.active FROM supplier_offers so JOIN supplier_profiles sp ON sp.id=so.supplier_id WHERE so.id=$1`, in.OfferID).Scan(&supplierID, &price, &stock, &minq, &status, &verification, &active)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "offer not found"})
		return
	}
	if status != "published" || verification != "verified" || !active || stock < in.Quantity || in.Quantity < minq {
		writeJSON(w, 409, map[string]string{"error": "offer is not eligible for this order"})
		return
	}
	if in.IdempotencyKey != "" {
		var existing string
		if err := s.DB.QueryRowContext(r.Context(), `SELECT id FROM orders WHERE idempotency_key=$1`, in.IdempotencyKey).Scan(&existing); err == nil && existing != orderID {
			writeJSON(w, 409, map[string]string{"error": "idempotency key already used"})
			return
		}
	}
	if _, err := s.DB.ExecContext(r.Context(), `UPDATE orders SET supplier_offer_id=$1,supplier_id=(SELECT organization_id FROM supplier_profiles WHERE id=$2),amount=amount+($3*$4) WHERE id=$5`, in.OfferID, supplierID, price, in.Quantity, orderID); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"order_id": orderID, "offer_id": in.OfferID, "quantity": in.Quantity, "unit_price": price, "supplier_id": supplierID})
}
