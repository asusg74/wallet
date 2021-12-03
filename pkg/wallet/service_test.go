package wallet

import (
	"testing"

	"github.com/google/uuid"
)

func TestReject_positive(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	payment, _ := svg.Pay(1, 50, "myWorld")
	svg.Reject(payment.ID)
	err := svg.Reject(payment.ID)
	if err != nil {
		t.Errorf("existing payment reject fail: %v", err)
	}
}
func TestReject_negative(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	svg.Pay(1, 50, "myWorld")
	err := svg.Reject("myCountry")
	if err != ErrPaymentNotFound {
		t.Errorf("missing payment reject success: %v", err)
	}
}

func TestRepeat_positive(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	payment, _ := svg.Pay(1, 50, "myWorld")
	_, err := svg.Repeat(payment.ID)
	if err != nil {
		t.Errorf("Repeat: %v", err)
	}
}

func TestRepeat_balanceLimit(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	payment, _ := svg.Pay(1, 60, "myWorld")
	_, err := svg.Repeat(payment.ID)
	if err != ErrNotEnoughBalance {
		t.Errorf("Repeat: %v", err)
	}
}

func TestRepeat_negative(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	_, err := svg.Repeat(uuid.New().String())
	if err != ErrPaymentNotFound {
		t.Errorf("Repeat: %v", err)
	}
}

func TestFavoritePayment_positive(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	payment, _ := svg.Pay(1, 60, "myWorld")
	_, err := svg.FavoritePayment(payment.ID, "smart city")
	if err != nil {
		t.Errorf("Set Favorite: %v", err)
	}
}

func TestFavoritePayment_negative(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	_, err := svg.FavoritePayment(uuid.New().String(), "smart city")
	if err == nil {
		t.Errorf("Set Favorite: %v", err)
	}
}

func TestPayFromFavorite_positive(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	pay, _ := svg.Pay(1, 40, "myWorld")
	favorite, _ := svg.FavoritePayment(pay.ID, "smart city")
	_, err := svg.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("Pay From Favorite: %v", err)
	}
}

func TestPayFromFavorite_wrongPay(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	_, err := svg.PayFromFavorite(uuid.New().String())
	if err == nil {
		t.Errorf("Pay From Favorite: %v", err)
	}
}

func TestPayFromFavorite_balanceLimit(t *testing.T) {
	svg := &Service{}
	svg.RegisterAccount("+992935007339")
	svg.Deposit(1, 100)
	pay, _ := svg.Pay(1, 60, "myWorld")
	favorite, _ := svg.FavoritePayment(pay.ID, "smart city")
	_, err := svg.PayFromFavorite(favorite.ID)
	if err == nil {
		t.Errorf("Pay From Favorite: %v", err)
	}
}

func TestFileworkAndEtc(t *testing.T) {
	svg := &Service{}
	account, _ := svg.RegisterAccount("+992935007339")
	svg.Deposit(account.ID, 100000)
	payment1, _ := svg.Pay(account.ID, 60, "smart city")
	payment2, _ := svg.Pay(account.ID, 120, "smart boy")
	payment3, _ := svg.Pay(account.ID, 300, "jackie")
	svg.FavoritePayment(payment1.ID, "boys1")
	svg.FavoritePayment(payment2.ID, "boys2")
	svg.FavoritePayment(payment3.ID, "boys3")
	err := svg.ExportToFile("gym.txt")
	if err != nil {
		t.Error(err)
	}
	err = svg.ImportFromFile("gym.txt")
	if err != nil {
		t.Error(err)
	}
	err = svg.Export("")
	if err != nil {
		t.Error(err)
	}
	err = svg.Import("")
	if err != nil {
		t.Error(err)
	}
}

