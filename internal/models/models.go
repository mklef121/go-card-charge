package models

import (
	"context"
	"database/sql"
	"time"
)

//Type for database connection values
type DBModel struct {
	DB *sql.DB
}

type Models struct {
	DB DBModel
}

//Returns a model type with database connection pool
func NewModel(db *sql.DB) Models {
	return Models{
		DB: DBModel{
			DB: db,
		},
	}
}

// Image          sql.NullString `json:"image"`
type Widget struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	InventoryLevel int       `json:"inventory_level"`
	Image          string    `json:"image"`
	Price          int       `json:"price"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

type Order struct {
	ID            int       `json:"id"`
	WidgetID      int       `json:"widget_id"`
	CustomerID    int       `json:"customer_id"`
	TransactionID int       `json:"transaction_id"`
	StatusID      int       `json:"status_id"`
	Quantity      int       `json:"quantity"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

type Status struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type TransactionStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Transaction struct {
	ID                  int       `json:"id"`
	TransactionStatusID int       `json:"transaction_status_id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	BankReturnCode      string    `json:"bank_return_code"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Customer struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (model *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := model.DB.QueryRowContext(ctx, `select 
	id, name, description, inventory_level, COALESCE(image, ''), price, created_at, updated_at
	from widgets 
	where id = ?`, id)

	err := row.Scan(&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Image,
		&widget.Price,
		&widget.CreatedAt,
		&widget.UpdatedAt)

	if err != nil {
		return widget, err
	}

	return widget, nil

}

func (model *DBModel) GetOrder(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var order Order

	row := model.DB.QueryRowContext(ctx, "select id,widget_id, transaction_id, status_id, quantity, amount from widgets where id = ?", id)

	err := row.Scan(&order.ID, &order.WidgetID, &order.TransactionID, &order.StatusID, &order.Quantity, &order.Amount)

	if err != nil {
		return order, err
	}

	return order, nil

}

func (model *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stm := `insert into transactions (
		transaction_status_id,
		amount,
		currency,
		last_four,
		bank_return_code,
		created_at,
		updated_at
	) values ( ?, ?, ?, ?, ?, ?, ?,)`

	result, err := model.DB.ExecContext(ctx,
		stm,
		txn.TransactionStatusID,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		time.Now(),
		time.Now())

	if err != nil {
		return 0, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastId), nil
}

func (model *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stm := `insert into orders (
		widget_id,
		transaction_id,
		status_id,
		quantity,
		amount,
		created_at,
		updated_at
	) values ( ?, ?, ?, ?, ?, ?, ?,)`

	result, err := model.DB.ExecContext(ctx,
		stm,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		time.Now(),
		time.Now())

	if err != nil {
		return 0, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastId), nil
}
