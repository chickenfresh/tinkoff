package tinkoff

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

type Notification struct {
	TerminalKey    string            `json:"TerminalKey"` // Идентификатор магазина
	OrderID        string            `json:"OrderId"`     // Номер заказа в системе Продавца
	Success        bool              `json:"Success"`     // Успешность операции
	Status         string            `json:"Status"`      // Статус платежа (см. описание статусов операций)
	PaymentID      uint64            `json:"PaymentId"`   // Уникальный идентификатор платежа. В случае нотификаций банк присылает число, а не строку, как в случае с Init или Cancel
	ErrorCode      string            `json:"ErrorCode"`   // Код ошибки, если произошла ошибка
	Amount         uint64            `json:"Amount"`      // Текущая сумма транзакции в копейках
	RebillID       uint64            `json:"RebillId"`    // Идентификатор рекуррентного платежа
	CardID         uint64            `json:"CardId"`      // Идентификатор привязанной карты
	PAN            string            `json:"Pan"`         // Маскированный номер карты
	DataStr        string            `json:"-"`
	Data           map[string]string `json:"-"`       // Дополнительные параметры платежа, переданные при создании заказа
	Token          string            `json:"Token"`   // Подпись запроса
	ExpirationDate string            `json:"ExpDate"` // Срок действия карты
}

type NotificationV1 struct {
	DataStr string `json:"DATA"`
}

type NotificationV2 struct {
	Data map[string]string `json:"data"` // Дополнительные параметры платежа, переданные при создании заказа
}

func (n *Notification) GetValuesForToken() map[string]string {
	var result = map[string]string{
		"TerminalKey": n.TerminalKey,
		"OrderId":     n.OrderID,
		"Success":     serializeBool(n.Success),
		"Status":      n.Status,
		"PaymentId":   strconv.FormatUint(n.PaymentID, 10),
		"ErrorCode":   n.ErrorCode,
		"Amount":      strconv.FormatUint(n.Amount, 10),
		"Pan":         n.PAN,
		"ExpDate":     n.ExpirationDate,
	}

	if n.CardID != 0 {
		result["CardId"] = strconv.FormatUint(n.CardID, 10)
	}

	if n.DataStr != "" {
		result["DATA"] = n.DataStr
	}

	if n.RebillID != 0 {
		result["RebillId"] = strconv.FormatUint(n.RebillID, 10)
	}
	return result
}

func (c *Client) ParseNotification(requestBody io.Reader) (*Notification, error) {
	bytes, err := ioutil.ReadAll(requestBody)
	if err != nil {
		return nil, err
	}
	var notification Notification
	err = json.Unmarshal(bytes, &notification)
	if err != nil {
		return nil, err
	}
	var dataV1 NotificationV1
	err = json.Unmarshal(bytes, &dataV1)
	if err != nil {
		var dataV2 NotificationV2
		err = json.Unmarshal(bytes, &dataV2)
		if err != nil {
			return nil, err
		}
		notification.Data = dataV2.Data
	} else if dataV1.DataStr != "" {
		notification.DataStr = dataV1.DataStr
		err = json.Unmarshal([]byte(dataV1.DataStr), &notification.Data)
		if err != nil {
			return nil, errors.New("can't unserialize DATA field: " + err.Error())
		}
	}

	if c.terminalKey != notification.TerminalKey {
		return nil, errors.New("invalid terminal key")
	}

	valuesForTokenGen := notification.GetValuesForToken()
	valuesForTokenGen["Password"] = c.password
	token := generateToken(valuesForTokenGen)
	if token != notification.Token {
		valsForTokenJSON, _ := json.Marshal(valuesForTokenGen)
		return nil, fmt.Errorf("invalid token: expected %s got %s.\nValues for token: %s.\nNotification: %s", token, notification.Token, valsForTokenJSON, string(bytes))
	}

	return &notification, nil
}

func (c *Client) GetNotificationSuccessResponse() string {
	return "OK"
}
