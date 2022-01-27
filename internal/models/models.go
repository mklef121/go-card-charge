package models

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	IsRecurring    bool      `json:"is_recurring"`
	StripePlanID   string    `json:"stripe_plan_id"`
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
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	BankReturnCode      string    `json:"bank_return_code"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
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
	id, 
	name, 
	description, 
	inventory_level, 
	COALESCE(image, ''), 
	price,
	is_recurring,
	stripe_plan_id,
	created_at, updated_at
	from widgets 
	where id = ?`, id)

	err := row.Scan(&widget.ID,
		&widget.Name,
		&widget.Description,
		&widget.InventoryLevel,
		&widget.Image,
		&widget.Price,
		&widget.IsRecurring,
		&widget.StripePlanID,
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
		expiry_month,
		expiry_year,
		payment_intent,
		payment_method,
		created_at,
		updated_at
	) values ( ?, ?, ?, ?, ?, ?, ?,?, ?,?,?)`

	result, err := model.DB.ExecContext(ctx,
		stm,
		txn.TransactionStatusID,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentIntent,
		txn.PaymentMethod,
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
		customer_id,
		created_at,
		updated_at
	) values ( ?, ?, ?, ?, ?, ?, ?,?)`

	result, err := model.DB.ExecContext(ctx,
		stm,
		order.WidgetID,
		order.TransactionID,
		order.StatusID,
		order.Quantity,
		order.Amount,
		order.CustomerID,
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

func (model *DBModel) InsertCustomer(cus Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stm := `insert into customers (
		first_name,
		last_name,
		email,
		created_at,
		updated_at
	) values ( ?, ?, ?, ?, ?)`

	result, err := model.DB.ExecContext(ctx,
		stm,
		cus.FirstName,
		cus.LastName,
		cus.Email,
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

func (model *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)

	row := model.DB.QueryRowContext(ctx, `select 
		id, first_name, last_name, email, password, created_at, updated_at  
		from users where email = ?`, email)

	var user User

	err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (model *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var passwordHash string

	res := model.DB.QueryRowContext(ctx, `select id, password from users where email = ?`, email)

	err := res.Scan(&id, &passwordHash)

	if err != nil {
		return id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))

	if err == bcrypt.ErrMismatchedHashAndPassword {

		return 0, errors.New("Incorrect password")
	} else if err != nil {
		return 0, err
	}

	return id, nil
}
