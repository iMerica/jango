package auth

import (
	"context"
	"fmt"

	"github.com/iMerica/jango/orm"
)

type ModelBackend struct{}

func (mb *ModelBackend) Authenticate(ctx context.Context, username, password string) (*User, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("auth: database not available")
	}

	userModel := GetUserModel()
	pkField := userModel.PKField
	passwordField := ""
	usernameField := ""

	for _, f := range userModel.Fields {
		if f.Name == "Username" || f.DBColumn == "username" {
			usernameField = f.DBColumn
			if usernameField == "" {
				usernameField = f.Name
			}
		}
		if f.Name == "Password" || f.DBColumn == "password" {
			passwordField = f.DBColumn
			if passwordField == "" {
				passwordField = f.Name
			}
		}
	}

	if usernameField == "" {
		usernameField = "username"
	}
	if passwordField == "" {
		passwordField = "password"
	}

	query := fmt.Sprintf("SELECT %s, %s, %s FROM %s WHERE %s = $1",
		pkField, usernameField, passwordField,
		userModel.TableName, usernameField)

	row := db.QueryRow(ctx, query, username)

	var id int64
	var uname, hashedPw string
	if err := row.Scan(&id, &uname, &hashedPw); err != nil {
		return nil, ErrAuthenticationFailed
	}

	if !CheckPassword(password, hashedPw) {
		return nil, ErrAuthenticationFailed
	}

	user := &User{
		ID:       id,
		Username: uname,
		Password: hashedPw,
	}

	if err := mb.loadUserFields(ctx, user, userModel, pkField, usernameField); err != nil {
		return user, nil
	}

	return user, nil
}

func (mb *ModelBackend) GetUser(ctx context.Context, userID int64) (*User, error) {
	db := orm.DefaultDB()
	if db == nil {
		return nil, fmt.Errorf("auth: database not available")
	}

	userModel := GetUserModel()

	cols := []string{}
	colNames := map[string]string{}
	for _, f := range userModel.Fields {
		col := f.DBColumn
		if col == "" {
			col = f.Name
		}
		if f.FieldType == orm.ManyToManyField {
			continue
		}
		cols = append(cols, col)
		colNames[f.Name] = col
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1",
		userModel.TableName, userModel.PKField)

	row := db.QueryRow(ctx, query, userID)
	record, err := row.Scan()
	if err != nil {
		return nil, ErrUserNotFound
	}

	user := &User{}
	if id, ok := record["id"]; ok {
		user.ID = id.(int64)
	}
	if username, ok := record["username"]; ok {
		user.Username = fmt.Sprintf("%v", username)
	}
	if email, ok := record["email"]; ok {
		user.Email = fmt.Sprintf("%v", email)
	}
	if password, ok := record["password"]; ok {
		user.Password = fmt.Sprintf("%v", password)
	}
	if isActive, ok := record["is_active"]; ok {
		user.IsActive = isActive.(bool)
	}
	if isStaff, ok := record["is_staff"]; ok {
		user.IsStaff = isStaff.(bool)
	}
	if isAdmin, ok := record["is_admin"]; ok {
		user.IsAdmin = isAdmin.(bool)
	}

	return user, nil
}

func (mb *ModelBackend) loadUserFields(ctx context.Context, user *User, userModel *orm.ModelMeta, pkCol, usernameCol string) error {
	db := orm.DefaultDB()
	if db == nil {
		return nil
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1",
		userModel.TableName, pkCol)

	row := db.QueryRow(ctx, query, user.ID)
	record, err := row.Scan()
	if err != nil {
		return err
	}

	if email, ok := record["email"]; ok {
		user.Email = fmt.Sprintf("%v", email)
	}
	if isActive, ok := record["is_active"]; ok {
		user.IsActive = toBool(isActive)
	}
	if isStaff, ok := record["is_staff"]; ok {
		user.IsStaff = toBool(isStaff)
	}
	if isAdmin, ok := record["is_admin"]; ok {
		user.IsAdmin = toBool(isAdmin)
	}

	return nil
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int64:
		return val != 0
	case string:
		return val == "true" || val == "1" || val == "t"
	default:
		return false
	}
}
