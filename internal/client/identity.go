package client

import (
	"time"
)

type registrationUnmarshal struct {
	Info struct {
		Display struct {
			Raw string
		}
	}
}

type AccountIdentity struct {
	Display string `json:"display,omitempty"`
}

func (c *Client) accountIdentityFromAccount(account []byte) (AccountIdentity, error) {
	var result registrationUnmarshal
	err := c.GetStorageRawWithTtl(
		"Identity",
		"IdentityOf",
		"Registration<BalanceOf>",
		6*time.Hour,
		&result,
		account,
	)
	if err != nil {
		return AccountIdentity{}, err
	}
	// Done
	r := AccountIdentity{
		Display: result.Info.Display.Raw,
	}
	return r, nil
}
