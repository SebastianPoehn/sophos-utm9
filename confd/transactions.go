package confd

// Transaction abstracts the read and write transactions to the confd
type Transaction interface {
	// Commit current transaction
	Commit() error
	// Rollback current Transaction
	Rollback() error
}

type writeTransaction struct{ *Conn }
type readTransaction struct{ *Conn }

// BeginReadTransaction starts new read transaction
func (c *Conn) BeginReadTransaction() (Transaction, error) {
	c.rwlock.RLock()
	_, err := c.SimpleRequest("freeze")
	if err != nil {
		return nil, err
	}
	return &readTransaction{c}, nil
}

// BeginWriteTransaction starts new write transaction
func (c *Conn) BeginWriteTransaction() (Transaction, error) {
	c.rwlock.Lock()
	_, err := c.SimpleRequest("lock")
	if err != nil {
		return nil, err
	}
	return &writeTransaction{c}, nil
}

func (t *readTransaction) Rollback() (err error) {
	return t.Commit()
}

func (t *readTransaction) Commit() (err error) {
	_, err = t.SimpleRequest("thaw")
	t.rwlock.RUnlock()
	return
}

func (t *writeTransaction) Rollback() (err error) {
	_, err = t.SimpleRequest("unlock")
	t.rwlock.Unlock()
	return
}

func (t *writeTransaction) Commit() (err error) {
	_, err = t.SimpleRequest("commit")
	t.rwlock.Unlock()
	return
}
