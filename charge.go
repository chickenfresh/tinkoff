package tinkoff

type ChargeRequest struct {
	BaseRequest
	PaymentID string `json:"PaymentId"`           // Идентификатор платежа в системе банка
	RebillId  string `json:"RebillId"`            // Идентификатор автоплатежа. Можно получить в нотификации AUTHORIZED или CONFIRMED родительского платежа
	SendEmail bool   `json:"SendEmail,omitempty"` // Получение покупателем уведомлений на электронную почту
	InfoEmail string `json:"InfoEmail,omitempty"` // Электронная почта покупателя
	ClientIP  string `json:"IP,omitempty"`        // IP-адрес покупателя
}

func (i *ChargeRequest) GetValuesForToken() map[string]string {
	v := map[string]string{
		"PaymentId": i.PaymentID,
		"RebillId":  i.RebillId,
		"IP":        i.ClientIP,
		"SendEmail": serializeBool(i.SendEmail),
		"InfoEmail": i.InfoEmail,
	}
	return v
}

type ChargeResponse struct {
	BaseResponse
	Amount    uint64 `json:"Amount"`    // Сумма в копейках
	OrderID   string `json:"OrderId"`   // Номер заказа в системе Продавца
	Status    string `json:"Status"`    // Статус транзакции
	PaymentID string `json:"PaymentId"` // Уникальный идентификатор транзакции в системе Банка. По офф. документации это number(20), но фактически значение передается в виде строки.
}

func (c *Client) Charge(request *ChargeRequest) (*ChargeResponse, error) {
	response, err := c.PostRequest("/Charge", request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var res ChargeResponse
	err = c.decodeResponse(response, &res)
	if err != nil {
		return nil, err
	}

	return &res, res.Error()
}
