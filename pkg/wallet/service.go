package wallet

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/asusg74/wallet/pkg/types"
)

var ErrPhoneRegistered = errors.New("phone is already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")
var ErrNotEnoughBalance = errors.New("account doesn't have enought balance")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, accounts := range s.accounts {
		if accounts.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)
	return account, nil
}

func (s Service) Deposit(accountID int64, amount types.Money) error {
	if amount < 0 {
		return ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return ErrAccountNotFound
	}
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount < 0 {
		return nil, ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if accountID == acc.ID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	var payment *types.Payment
	for _, pay := range s.payments {
		if pay.ID == paymentID {
			payment = pay
			break
		}
	}
	if payment == nil {
		return nil, ErrPaymentNotFound
	}
	return payment, nil
}
func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	var favorite *types.Favorite
	for _, fav := range s.favorites {
		if fav.ID == favoriteID {
			favorite = fav
			break
		}
	}
	if favorite == nil {
		return nil, ErrFavoriteNotFound
	}
	return favorite, nil
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	payment.Status = types.PaymentStatusFail
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}
	account.Balance += payment.Amount
	return nil
}
func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	currentPayment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	newPayment, err := s.Pay(currentPayment.AccountID, currentPayment.Amount, currentPayment.Category)
	if err != nil {
		return nil, err
	}
	return newPayment, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	result := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}
	s.favorites = append(s.favorites, result)
	return result, nil
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}
	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Println("can't create file")
		return err
	}
	str := ""
	for _, account := range s.accounts {
		str += strconv.FormatInt(account.ID, 10) + ";"
		str += string(account.Phone) + ";"
		str += strconv.FormatInt(int64(account.Balance), 10)
		str += "|"
	}
	_, err = file.Write([]byte(str))
	if err != nil {
		log.Println("can't write file str:", str)
		return err
	}

	cerr := file.Close()
	if cerr != nil {
		log.Println("can't close file")
		return cerr
	}
	return nil
}

func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	content := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		read, err := file.Read(buf)
		if err == io.EOF {
			content = append(content, buf[:read]...)
			break
		}
		if err != nil {
			return err
		}
		content = append(content, buf[:read]...)
	}
	data := strings.Split(string(content), "|")
	for _, accountDt := range data {
		fields := strings.Split(accountDt, ";")
		if len(fields) < 3 {
			continue
		}
		ID, _ := strconv.ParseInt(fields[0], 10, 64)
		balance, _ := strconv.ParseInt(fields[2], 10, 64)
		s.accounts = append(s.accounts, &types.Account{
			ID:      ID,
			Phone:   types.Phone(fields[1]),
			Balance: types.Money(balance),
		})
	}

	return nil
}

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (s *Service) ExportAccountsToString(sep1 string, sep2 string) string {
	str := ""
	for _, account := range s.accounts {
		str += strconv.FormatInt(account.ID, 10) + sep1
		str += string(account.Phone) + sep1
		str += strconv.FormatInt(int64(account.Balance), 10)
		str += sep2
	}
	return str
}
func (s *Service) ExportPaymentsToString(sep1 string, sep2 string) string {
	str := ""
	for _, payment := range s.payments {
		str += payment.ID + sep1
		str += strconv.FormatInt(payment.AccountID, 10) + sep1
		str += strconv.FormatInt(int64(payment.Amount), 10) + sep1
		str += string(payment.Category) + sep1
		str += string(payment.Status) + sep1
		str += sep2
	}
	return str
}
func (s *Service) ExportFavoritesToString(sep1 string, sep2 string) string {
	str := ""
	for _, favorite := range s.favorites {
		str += favorite.ID + sep1
		str += strconv.FormatInt(favorite.AccountID, 10) + sep1
		str += string(favorite.Name) + sep1
		str += strconv.FormatInt(int64(favorite.Amount), 10) + sep1
		str += string(favorite.Category) + sep1
		str += sep2
	}
	return str
}

func ImportAccounts(path string, sep1, sep2 string) ([]*types.Account, error) {
	b, _ := Exists(path)
	var accounts []*types.Account
	if b {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		content := make([]byte, 0)
		buf := make([]byte, 4096)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				content = append(content, buf[:read]...)
				break
			}
			if err != nil {
				return nil, err
			}
			content = append(content, buf[:read]...)
		}
		data := strings.Split(string(content), sep2)
		for _, accountDt := range data {
			fields := strings.Split(accountDt, sep1)
			if len(fields) < 3 {
				continue
			}
			ID, _ := strconv.ParseInt(fields[0], 10, 64)
			balance, _ := strconv.ParseInt(fields[2], 10, 64)
			accounts = append(accounts, &types.Account{
				ID:      ID,
				Phone:   types.Phone(fields[1]),
				Balance: types.Money(balance),
			})
		}
	}
	return accounts, nil
}
func ImportPayments(path string, sep1, sep2 string) ([]*types.Payment, error) {
	b, _ := Exists(path)
	var payments []*types.Payment
	if b {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		content := make([]byte, 0)
		buf := make([]byte, 4096)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				content = append(content, buf[:read]...)
				break
			}
			if err != nil {
				return nil, err
			}
			content = append(content, buf[:read]...)
		}
		data := strings.Split(string(content), sep2)
		for _, paymentDt := range data {
			fields := strings.Split(paymentDt, sep1)
			if len(fields) < 5 {
				continue
			}
			accID, _ := strconv.ParseInt(fields[1], 10, 64)
			amount, _ := strconv.ParseInt(fields[2], 10, 64)
			payments = append(payments, &types.Payment{
				ID:        fields[0],
				AccountID: accID,
				Amount:    types.Money(amount),
				Category:  types.PaymentCategory(fields[3]),
				Status:    types.PaymentStatus(fields[4]),
			})
		}
	}
	return payments, nil
}
func ImportFavorites(path string, sep1, sep2 string) ([]*types.Favorite, error) {
	b, _ := Exists(path)
	var favorites []*types.Favorite
	if b {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		content := make([]byte, 0)
		buf := make([]byte, 4096)
		for {
			read, err := file.Read(buf)
			if err == io.EOF {
				content = append(content, buf[:read]...)
				break
			}
			if err != nil {
				return nil, err
			}
			content = append(content, buf[:read]...)
		}
		data := strings.Split(string(content), sep2)
		for _, favoriteDt := range data {
			fields := strings.Split(favoriteDt, sep1)
			if len(fields) < 5 {
				continue
			}
			accID, _ := strconv.ParseInt(fields[1], 10, 64)
			amount, _ := strconv.ParseInt(fields[3], 10, 64)
			favorites = append(favorites, &types.Favorite{
				ID:        fields[0],
				AccountID: accID,
				Name:      fields[2],
				Amount:    types.Money(amount),
				Category:  types.PaymentCategory(fields[4]),
			})
		}
	}
	return favorites, nil
}

func (s *Service) Export(dir string) error {
	dir = filepath.Clean(dir)
	if len(s.accounts) != 0 {
		filePath := dir + "/accounts.dump"
		_, err := os.Create(filePath)
		if err != nil {
			return err
		}
		err = os.WriteFile(filePath, []byte(s.ExportAccountsToString(";", "\n")), 0666)
		if err != nil {
			return err
		}
	}
	if len(s.payments) != 0 {
		filePath := dir + "/payments.dump"
		_, err := os.Create(filePath)
		if err != nil {
			return err
		}
		err = os.WriteFile(filePath, []byte(s.ExportPaymentsToString(";", "\n")), 0666)
		if err != nil {
			return err
		}
	}
	if len(s.favorites) != 0 {
		filePath := dir + "/favorites.dump"
		_, err := os.Create(filePath)
		if err != nil {
			return err
		}
		err = os.WriteFile(filePath, []byte(s.ExportFavoritesToString(";", "\n")), 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Import(dir string) error {
	dir = filepath.Clean(dir)
	newAcc, err := ImportAccounts(dir+"/accounts.dump", ";", "\n")
	if err != nil {
		return err
	}
	for _, acc := range newAcc {
		existAcc, err := s.FindAccountByID(acc.ID)
		if err == ErrAccountNotFound {
			s.accounts = append(s.accounts, acc)
			s.nextAccountID++
			continue
		}
		existAcc.Balance = acc.Balance
		existAcc.Phone = acc.Phone
	}

	newPay, err := ImportPayments(dir+"/payments.dump", ";", "\n")
	if err != nil {
		return err
	}
	for _, pay := range newPay {
		existPay, err := s.FindPaymentByID(pay.ID)
		if err == ErrPaymentNotFound {
			s.payments = append(s.payments, pay)
			continue
		}
		existPay.AccountID = pay.AccountID
		existPay.Amount = pay.Amount
		existPay.Category = pay.Category
		existPay.Status = pay.Status
	}

	newFav, err := ImportFavorites(dir+"/favorites.dump", ";", "\n")
	if err != nil {
		return err
	}
	for _, fav := range newFav {
		existFav, err := s.FindFavoriteByID(fav.ID)
		if err == ErrFavoriteNotFound {
			s.favorites = append(s.favorites, fav)
			continue
		}
		existFav.AccountID = fav.AccountID
		existFav.Name = fav.Name
		existFav.Amount = fav.Amount
		existFav.Category = fav.Category
	}
	return nil
}
