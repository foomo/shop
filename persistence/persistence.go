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

// NewPersistorWithIndexes will construct a new persistor and ensure indices
func NewPersistorWithIndexes(mongoURL string, collectionName string, indexes []mgo.Index) (p *Persistor, err error) {

	// construct persitor
	p, err = NewPersistor(mongoURL, collectionName)
	if err != nil {
		return
	}

	// get collection
	session, collection := p.GetCollection()
	defer session.Close()

	// iterate indexes
	for _, index := range indexes {

		// ensure index
		indexErr := collection.EnsureIndex(index)
		if indexErr != nil {
			return nil, indexErr
		}

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

// GetGlobalSessionCollection is used when multiple threads share the same connections (bad idea)
// and should be used ONLY when necessary. Use get collection and return the connection to the connection
// pool instead by invoking session.close() instead
func (p *Persistor) GetGlobalSessionCollection() (collection *mgo.Collection) {
	if err := p.session.Ping(); err != nil {
		p.session.Refresh()
	}
	return p.session.DB(p.db).C(p.collection)
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
