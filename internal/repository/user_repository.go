package repository

import (
	"database/sql"
	"errors"
	"messenger/internal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *model.User) error {
	query := `INSERT INTO users(username, email, password, phone) VALUES($1,$2,$3,$4) RETURNING id, created_at;`

	var email interface{} = u.Email
	if u.Email == "" {
		email = nil
	}
	var phone interface{} = u.Phone
	if u.Phone == "" {
		phone = nil
	}

	return r.db.QueryRow(query, u.Username, email, u.Password, phone).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status, profession sql.NullString
	var birthDate sql.NullTime
	err := r.db.QueryRow(
		"SELECT id, username, email, password, phone, full_name, birth_date, location, status, COALESCE(avatar_url,''), profession FROM users WHERE email = $1", email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.Password, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl, &profession)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	u.Profession = profession.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

func (r *UserRepository) GetById(id uuid.UUID) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status, profession, email sql.NullString
	var birthDate sql.NullTime
	err := r.db.QueryRow(
		"SELECT id, username, email, phone, full_name, birth_date, location, status, COALESCE(avatar_url,''), profession FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.Username, &email, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl, &profession)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Email = email.String
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	u.Profession = profession.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

func (r *UserRepository) GetByUsernameAndEmail(username, email, phone string) (*model.User, error) {
	var u model.User
	var e, p sql.NullString
	query := `SELECT id, username, email, phone FROM users WHERE username = $1 
	          OR (email = $2 AND email != '') 
	          OR (phone = $3 AND phone != '') 
	          LIMIT 1`
	err := r.db.QueryRow(query, username, email, phone).Scan(&u.ID, &u.Username, &e, &p)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Email = e.String
	u.Phone = p.String
	return &u, nil
}

func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status, profession, email sql.NullString
	var birthDate sql.NullTime
	query := `SELECT id, username, email, phone, full_name, birth_date, location, status, COALESCE(avatar_url,''), profession FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &email, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl, &profession)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Email = email.String
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	u.Profession = profession.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

func (r *UserRepository) SearchByUsername(username string) ([]model.User, error) {
	q := "%" + username + "%"
	query := `SELECT id, username, COALESCE(email,''), COALESCE(avatar_url,'')
	          FROM users
	          WHERE username ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1
	          LIMIT 10`
	rows, err := r.db.Query(query, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.AvatarUrl); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) VerifyPassword(email, password string) (*model.User, error) {
	user, err := r.GetByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("пользователь не найден")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	user.Password = ""
	return user, nil
}

func (r *UserRepository) UpdateProfile(u *model.User) error {
	query := `UPDATE users SET phone = $1, full_name = $2, birth_date = $3, location = $4, status = $5, username = $6, profession = $7, updated_at = NOW() WHERE id = $8`
	// Пустые строки сохраняем как NULL в БД
	locationVal := sql.NullString{String: u.Location, Valid: u.Location != ""}
	professionVal := sql.NullString{String: u.Profession, Valid: u.Profession != ""}
	phoneVal := sql.NullString{String: u.Phone, Valid: u.Phone != ""}
	fullNameVal := sql.NullString{String: u.FullName, Valid: u.FullName != ""}
	_, err := r.db.Exec(query, phoneVal, fullNameVal, u.BirthDate, locationVal, u.Status, u.Username, professionVal, u.ID)
	return err
}

func (r *UserRepository) UpdateStatus(id uuid.UUID, status string) (*model.User, error) {
	query := `UPDATE users SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return nil, err
	}

	return r.GetById(id)
}

func (r *UserRepository) UpdateAvatarUrl(userID uuid.UUID, url string) error {
	_, err := r.db.Exec(`UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2`, url, userID)
	return err
}

// GetByPhone ищет пользователя по номеру телефона
func (r *UserRepository) GetByPhone(phone string) (*model.User, error) {
	u := new(model.User)
	var phoneNull, fullName, location, status, profession, email sql.NullString
	var birthDate sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, username, email, password, phone, full_name, birth_date, location, status, COALESCE(avatar_url,''), profession 
		 FROM users WHERE phone = $1`, phone,
	).Scan(&u.ID, &u.Username, &email, &u.Password, &phoneNull, &fullName, &birthDate, &location, &status, &u.AvatarUrl, &profession)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Email = email.String
	u.Phone = phoneNull.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	u.Profession = profession.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}

// CreateByPhone создаёт пользователя только с телефоном (без пароля и email)
func (r *UserRepository) CreateByPhone(phone, username string) (*model.User, error) {
	u := &model.User{
		Phone:    phone,
		Username: username,
	}
	err := r.db.QueryRow(
		`INSERT INTO users(username, phone) VALUES($1, $2) RETURNING id, created_at`,
		username, phone,
	).Scan(&u.ID, &u.CreatedAt)
	return u, err
}

func (r *UserRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE users SET password=$1, updated_at=NOW() WHERE id=$2`, hashedPassword, userID)
	return err
}

func (r *UserRepository) GetByEmailOrPhone(login string) (*model.User, error) {
	u := new(model.User)
	var phone, fullName, location, status, profession, email sql.NullString
	var birthDate sql.NullTime
	query := `SELECT id, username, email, phone, full_name, birth_date, location, status, COALESCE(avatar_url,''), profession
	          FROM users WHERE email=$1 OR phone=$1 LIMIT 1`
	err := r.db.QueryRow(query, login).Scan(&u.ID, &u.Username, &email, &phone, &fullName, &birthDate, &location, &status, &u.AvatarUrl, &profession)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Email = email.String
	u.Phone = phone.String
	u.FullName = fullName.String
	u.Location = location.String
	u.Status = status.String
	u.Profession = profession.String
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	return u, nil
}
