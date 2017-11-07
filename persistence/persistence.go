package persistence

import (
	"errors"
	"fmt"
	"net/url"

	"gopkg.in/mgo.v2"
)

//------------------------------------------------------------------
// ~ PUBLIC TYPES
//------------------------------------------------------------------

// Persistor persist
type Persistor struct {
	session    *mgo.Session
	url        string
	db         string
	collection string
}

//------------------------------------------------------------------
// ~ CONSTRUCTOR
//------------------------------------------------------------------

func NewPersistorWithIndex(mongoURL string, collectionName string, index mgo.Index) (p *Persistor, err error) {
	p, err = NewPersistor(mongoURL, collectionName)
	if err != nil {
		return
	}

	session, collection := p.GetCollection()
	defer session.Close()

	err = collection.EnsureIndex(index)
	if err != nil {
		return
	}
	return p, nil
}

// NewPersistor constructor
func NewPersistor(mongoURL string, collection string) (p *Persistor, err error) {
	parsedURL, err := url.Parse(mongoURL)
	if err != nil {
		return nil, err
	}
	if parsedURL.Scheme != "mongodb" {
		return nil, fmt.Errorf("missing scheme mongo:// in %q", mongoURL)
	}
	if len(parsedURL.Path) < 2 {
		return nil, errors.New("invalid mongoURL missing db should be mongodb://server:port/db")
	}
	session, err := mgo.Dial(mongoURL)
	if err != nil {
		return nil, err
	}
	p = &Persistor{
		session:    session,
		url:        mongoURL,
		db:         parsedURL.Path[1:],
		collection: collection,
	}
	return p, nil
}

//------------------------------------------------------------------
// ~ PUBLIC METHODS
//------------------------------------------------------------------

func (p *Persistor) GetCollection() (session *mgo.Session, collection *mgo.Collection) {
	session = p.session.Clone()
	collection = session.DB(p.db).C(p.collection)
	return session, collection
}

func (p *Persistor) GetCollectionName() string {
	return p.collection
}
func (p *Persistor) GetURL() string {
	return p.url
}
func (p *Persistor) LockDb() error {
	return p.session.FsyncLock()
}
func (p *Persistor) UnlockDb() error {
	return p.session.FsyncUnlock()
}
