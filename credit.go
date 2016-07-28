package smb2

import (
	"sync"
)

type account struct {
	c                *sync.Cond
	balance          uint16
	_opening         uint16
	maxCreditBalance uint16
}

func openAccount(maxCreditBalance uint16) *account {
	return &account{
		c:                sync.NewCond(new(sync.Mutex)),
		balance:          1,
		maxCreditBalance: maxCreditBalance,
	}
}

func (a *account) initRequest() uint16 {
	return a.maxCreditBalance - a.balance
}

func (a *account) request(creditCharge uint16) (uint16, bool) {
	if creditCharge == 0 {
		return 0, true
	}

	a.c.L.Lock()
	defer a.c.L.Unlock()

	for a.balance == 0 {
		a.c.Wait()
	}

	if a.balance < creditCharge {
		creditCharge = a.balance
		a.balance = 0
		return creditCharge, false
	}

	a.balance -= creditCharge

	return creditCharge, true
}

func (a *account) opening() uint16 {
	a.c.L.Lock()
	defer a.c.L.Unlock()

	ret := a._opening
	a._opening = 0

	return ret
}

func (a *account) grant(granted, requested uint16) {
	if granted == 0 && requested == 0 {
		return
	}

	a.c.L.Lock()
	defer a.c.Broadcast()
	defer a.c.L.Unlock()

	if granted < requested {
		a._opening += requested - granted
	}
	a.balance += granted
}