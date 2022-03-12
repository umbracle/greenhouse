package state

import (
	"fmt"

	"github.com/hashicorp/go-memdb"
)

type Handle struct {
	Dir      string
	Filename string
}

type Document struct {
	Dir      string
	Filename string
	Text     []byte
}

func (d *Document) Copy() *Document {
	return &Document{
		Dir:      d.Dir,
		Filename: d.Filename,
		Text:     d.Text,
	}
}

type State struct {
	db *memdb.MemDB
}

func NewState() *State {
	memdb, err := memdb.NewMemDB(dbSchema)
	if err != nil {
		panic(err)
	}
	s := &State{
		db: memdb,
	}
	return s
}

func (s *State) OpenDocument(dh Handle, text []byte) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	obj, err := txn.First(documentsTableName, "id", dh.Dir, dh.Filename)
	if err != nil {
		return err
	}
	if obj != nil {
		return fmt.Errorf("file already exists")
	}

	doc := &Document{
		Dir:      dh.Dir,
		Filename: dh.Filename,
		Text:     text,
	}

	err = txn.Insert(documentsTableName, doc)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *State) UpdateDocument(dh Handle, newText []byte) error {
	txn := s.db.Txn(true)
	defer txn.Abort()

	doc, err := s.getDocumentImpl(txn, dh)
	if err != nil {
		return err
	}

	doc = doc.Copy()
	doc.Text = newText

	err = txn.Insert(documentsTableName, doc)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (s *State) GetDocument(dh Handle) (*Document, error) {
	txn := s.db.Txn(false)
	return s.getDocumentImpl(txn, dh)
}

func (s *State) getDocumentImpl(txn *memdb.Txn, dh Handle) (*Document, error) {
	obj, err := txn.First(documentsTableName, "id", dh.Dir, dh.Filename)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, fmt.Errorf("doc not found")
	}
	return obj.(*Document), nil
}
