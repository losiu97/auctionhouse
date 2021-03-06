package dao

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"golang-poc/dto"
	"golang.org/x/crypto/bcrypt"
	"os"
)

type Token struct {
	UserId uint
	jwt.StandardClaims
}

type Account struct {
	CommonModelFields
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token";sql:"-"`
}

func (account *Account) ValidateUnique() (string, bool) {
	temp := &Account{}

	err := GetDB().Table("accounts").Where("email = ?", account.Email).First(temp).Error

	if account.failedToGetRecord(err) {
		return "Connection error. Please retry", false
	}

	if temp.Email != "" {
		return "Email address already in use by another user.", false
	}

	return "Requirement passed", true
}

func Create(createAccountDto *dto.CreateAccountDto) (*Account, error) {
	account := NewAccount(createAccountDto)

	if resp, ok := account.ValidateUnique(); !ok {
		return nil, errors.New(resp)
	}

	err := GetDB().Create(account).Error

	if err != nil {
		return nil, err
	}

	account.generateJwtToken()

	return account, nil
}

func Login(email, password string) (*Account, error) {

	account := &Account{}
	err := GetDB().Table("accounts").Where("email = ?", email).First(account).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, errors.New("invalid login credentials")
	}

	account.generateJwtToken()

	return account, nil
}

func GetUser(u uint) (*Account, error) {

	acc := &Account{}
	err := GetDB().Table("accounts").Where("id = ?", u).First(acc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("account not found")
		}
		return nil, err
	}

	return acc, nil
}

func (account *Account) generateJwtToken() {
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString
}

func (account *Account) failedToGetRecord(err error) bool {
	return err != nil && err != gorm.ErrRecordNotFound
}

func NewAccount(accountDto *dto.CreateAccountDto) *Account {
	account := &Account{}
	account.Email = accountDto.Email
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(accountDto.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	return account
}
