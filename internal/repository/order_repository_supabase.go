package repository

import (
	"context"
	"fmt"
	"time"

	"chatbot-automation/internal/database"
	"chatbot-automation/internal/models"

	"github.com/sirupsen/logrus"
)

// orderRepositorySupabase implements OrderRepository using Supabase SDK
type orderRepositorySupabase struct {
	supabase *database.SupabaseSDK
}

// NewOrderRepositorySupabase creates a Supabase-based order repository
func NewOrderRepositorySupabase(supabase *database.SupabaseSDK) OrderRepository {
	return &orderRepositorySupabase{supabase: supabase}
}

// CreateOrder creates a new order in the database
func (r *orderRepositorySupabase) CreateOrder(order *models.Order) (int, error) {
	ctx := context.Background()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	var result models.Order
	err := r.supabase.From("orders").Insert(ctx, order, &result)
	if err != nil {
		return 0, fmt.Errorf("failed to create order via Supabase: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"order_id": result.ID,
		"amount":   order.Amount,
		"method":   order.Method,
	}).Info("Order created successfully via Supabase")

	return result.ID, nil
}

// GetOrderByID retrieves an order by ID
func (r *orderRepositorySupabase) GetOrderByID(id int) (*models.Order, error) {
	ctx := context.Background()

	var orders []models.Order
	err := r.supabase.From("orders").
		Select("*").
		Eq("id", id).
		Limit(1).
		Execute(ctx, &orders)

	if err != nil {
		return nil, fmt.Errorf("failed to get order via Supabase: %w", err)
	}

	if len(orders) == 0 {
		return nil, nil
	}

	return &orders[0], nil
}

// GetOrderByBillID retrieves an order by Billplz bill ID
func (r *orderRepositorySupabase) GetOrderByBillID(billID string) (*models.Order, error) {
	ctx := context.Background()

	var orders []models.Order
	err := r.supabase.From("orders").
		Select("*").
		Eq("bill_id", billID).
		Limit(1).
		Execute(ctx, &orders)

	if err != nil {
		return nil, fmt.Errorf("failed to get order by bill_id via Supabase: %w", err)
	}

	if len(orders) == 0 {
		return nil, nil
	}

	return &orders[0], nil
}

// UpdateOrderStatus updates the payment status of an order
func (r *orderRepositorySupabase) UpdateOrderStatus(billID string, status string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("orders").
		Eq("bill_id", billID).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update order status via Supabase: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"bill_id": billID,
		"status":  status,
	}).Info("Order status updated successfully via Supabase")

	return nil
}

// UpdateOrderBillInfo updates the Billplz bill ID and URL for an order
func (r *orderRepositorySupabase) UpdateOrderBillInfo(orderID int, billID string, url string) error {
	ctx := context.Background()

	updateData := map[string]interface{}{
		"bill_id":    billID,
		"url":        url,
		"updated_at": time.Now(),
	}

	err := r.supabase.From("orders").
		Eq("id", orderID).
		Update(ctx, updateData)

	if err != nil {
		return fmt.Errorf("failed to update order bill info via Supabase: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"bill_id":  billID,
	}).Info("Order bill info updated successfully via Supabase")

	return nil
}

// GetOrdersByUserID retrieves orders for a specific user
func (r *orderRepositorySupabase) GetOrdersByUserID(userID string, limit int, offset int) ([]models.Order, int, error) {
	ctx := context.Background()

	// Get all orders for count (Supabase REST API doesn't have direct COUNT support)
	var allOrders []models.Order
	err := r.supabase.From("orders").
		Select("id").
		Eq("user_id", userID).
		Execute(ctx, &allOrders)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count via Supabase: %w", err)
	}

	totalCount := len(allOrders)

	// Get paginated orders
	var orders []models.Order
	err = r.supabase.From("orders").
		Select("*").
		Eq("user_id", userID).
		Order("created_at", false). // descending
		Limit(limit).
		Offset(offset).
		Execute(ctx, &orders)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get orders via Supabase: %w", err)
	}

	return orders, totalCount, nil
}

// GetAllOrders retrieves all orders (admin use)
func (r *orderRepositorySupabase) GetAllOrders(limit int, offset int) ([]models.Order, int, error) {
	ctx := context.Background()

	// Get all orders for count
	var allOrders []models.Order
	err := r.supabase.From("orders").
		Select("id").
		Execute(ctx, &allOrders)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count via Supabase: %w", err)
	}

	totalCount := len(allOrders)

	// Get paginated orders
	var orders []models.Order
	err = r.supabase.From("orders").
		Select("*").
		Order("created_at", false). // descending
		Limit(limit).
		Offset(offset).
		Execute(ctx, &orders)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get all orders via Supabase: %w", err)
	}

	return orders, totalCount, nil
}
