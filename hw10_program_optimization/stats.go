package hw10programoptimization

import (
	"bufio"
	"errors"
	"io"
	"strings"

	"github.com/valyala/fastjson"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	domain = "." + domain
	stat := make(DomainStat)
	err := parsUsers(
		r,
		func(u User) error {
			if strings.HasSuffix(u.Email, domain) {
				parts := strings.SplitN(u.Email, "@", 2)
				if len(parts) != 2 {
					return errors.New("incorrect email: " + u.Email)
				}
				key := strings.ToLower(parts[1])
				stat[key]++
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return stat, nil
}

func parsUsers(r io.Reader, handler func(user User) error) error {
	scanner := bufio.NewScanner(r)
	var value *fastjson.Value
	var p fastjson.Parser
	for scanner.Scan() {
		var err error
		if value, err = p.Parse(scanner.Text()); err != nil {
			return err
		}

		if err := handler(User{
			ID:       value.GetInt("Id"),
			Name:     string(value.GetStringBytes("Name")),
			Username: string(value.GetStringBytes("Username")),
			Email:    string(value.GetStringBytes("Email")),
			Phone:    string(value.GetStringBytes("Phone")),
			Password: string(value.GetStringBytes("Password")),
			Address:  string(value.GetStringBytes("Address")),
		}); err != nil {
			return err
		}
	}
	return nil
}
