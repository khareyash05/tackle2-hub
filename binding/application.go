package binding

import (
	"bytes"
	"errors"
	mime "github.com/gin-gonic/gin/binding"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/api"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"strconv"
)

//
// Application API.
type Application struct {
	client *Client
}

//
// Create an Application.
func (h *Application) Create(r *api.Application) (err error) {
	err = h.client.Post(api.ApplicationsRoot, &r)
	return
}

//
// Get an Application by ID.
func (h *Application) Get(id uint) (r *api.Application, err error) {
	r = &api.Application{}
	path := Path(api.ApplicationRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

//
// List Applications.
func (h *Application) List() (list []api.Application, err error) {
	list = []api.Application{}
	err = h.client.Get(api.ApplicationsRoot, &list)
	return
}

//
// Update an Application.
func (h *Application) Update(r *api.Application) (err error) {
	path := Path(api.ApplicationRoot).Inject(Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

//
// Delete an Application.
func (h *Application) Delete(id uint) (err error) {
	err = h.client.Delete(Path(api.ApplicationRoot).Inject(Params{api.ID: id}))
	return
}

//
// Bucket returns the bucket API.
func (h *Application) Bucket(id uint) (b *BucketContent) {
	params := Params{
		api.Wildcard: "",
		api.ID:       id,
	}
	path := Path(api.AppBucketContentRoot).Inject(params)
	b = &BucketContent{
		root:   path,
		client: h.client,
	}
	return
}

//
// FindIdentity by kind.
func (h *Application) FindIdentity(id uint, kind string) (r *api.Identity, found bool, err error) {
	list := []api.Identity{}
	p1 := Param{
		Key:   api.AppId,
		Value: strconv.Itoa(int(id)),
	}
	p2 := Param{
		Key:   api.Decrypted,
		Value: "1",
	}
	path := Path(api.IdentitiesRoot).Inject(Params{api.ID: id})
	err = h.client.Get(path, &list, p1, p2)
	if err != nil {
		return
	}
	for i := range list {
		r = &list[i]
		if r.Kind == kind {
			m := r.Model()
			r.With(m)
			found = true
			break
		}
	}
	return
}

//
// Tags returns the tags API.
func (h *Application) Tags(id uint) (tg AppTags) {
	tg = AppTags{
		client: h.client,
		appId:  id,
	}
	return
}

//
// AppTags sub-resource API.
// Provides association management of tags to applications by name.
type AppTags struct {
	client *Client
	appId  uint
	source *string
}

//
// Source sets the source for other operations on the associated tags.
func (h *AppTags) Source(name string) {
	h.source = &name
}

//
// Replace the associated tags for the source with a new set.
// Returns an error if the source is not set.
func (h *AppTags) Replace(ids []uint) (err error) {
	if h.source == nil {
		err = liberr.New("Source required.")
		return
	}
	path := Path(api.ApplicationTagsRoot).Inject(Params{api.ID: h.appId})
	query := []Param{}
	if h.source != nil {
		query = append(
			query,
			Param{
				Key:   api.Source,
				Value: *h.source,
			})
	}
	tags := []api.TagRef{}
	for _, id := range h.unique(ids) {
		tags = append(tags, api.TagRef{ID: id})
	}
	err = h.client.Put(path, tags, query...)
	return
}

//
// List associated tags.
// Returns a list of tag names.
func (h *AppTags) List() (list []api.TagRef, err error) {
	list = []api.TagRef{}
	path := Path(api.ApplicationTagsRoot).Inject(Params{api.ID: h.appId})
	query := []Param{}
	if h.source != nil {
		query = append(
			query,
			Param{
				Key:   api.Source,
				Value: *h.source,
			})
	}
	err = h.client.Get(path, &list, query...)
	return
}

//
// Add associates a tag with the application.
func (h *AppTags) Add(id uint) (err error) {
	path := Path(api.ApplicationTagsRoot).Inject(Params{api.ID: h.appId})
	tag := api.TagRef{ID: id}
	if h.source != nil {
		tag.Source = *h.source
	}
	err = h.client.Post(path, &tag)
	return
}

//
// Ensure ensures tag is associated with the application.
func (h *AppTags) Ensure(id uint) (err error) {
	err = h.Add(id)
	if errors.Is(err, &Conflict{}) {
		err = nil
	}
	return
}

//
// Delete ensures the tag is not associated with the application.
func (h *AppTags) Delete(id uint) (err error) {
	path := Path(
		api.ApplicationTagRoot).Inject(
		Params{
			api.ID:  h.appId,
			api.ID2: id})
	query := []Param{}
	if h.source != nil {
		query = append(query, Param{Key: api.Source, Value: *h.source})
	}
	err = h.client.Delete(path, query...)
	return
}

//
// unique ensures unique ids.
func (h *AppTags) unique(ids []uint) (u []uint) {
	mp := map[uint]int{}
	for _, id := range ids {
		mp[id] = 0
	}
	for id := range mp {
		u = append(u, id)
	}
	return
}

//
// Facts returns the tags API.
func (h *Application) Facts(id uint) (f AppFacts) {
	f = AppFacts{
		client: h.client,
		appId:  id,
	}
	return
}

//
// AppFacts sub-resource API.
// Provides association management of facts.
type AppFacts struct {
	client *Client
	appId  uint
	source string
}

//
// Source sets the source for other operations on the facts.
func (h *AppFacts) Source(source string) {
	h.source = source
}

//
// List facts.
func (h *AppFacts) List() (facts api.FactMap, err error) {
	facts = api.FactMap{}
	key := api.FactKey("")
	key.Qualify(h.source)
	path := Path(api.ApplicationFactsRoot).Inject(Params{api.ID: h.appId, api.Key: key})
	err = h.client.Get(path, &facts)
	return
}

//
// Get a fact.
func (h *AppFacts) Get(name string, value interface{}) (err error) {
	key := api.FactKey(name)
	key.Qualify(h.source)
	path := Path(api.ApplicationFactRoot).Inject(
		Params{
			api.ID:  h.appId,
			api.Key: key,
		})
	err = h.client.Get(path, value)
	return
}

//
// Set a fact (created as needed).
func (h *AppFacts) Set(name string, value interface{}) (err error) {
	key := api.FactKey(name)
	key.Qualify(h.source)
	path := Path(api.ApplicationFactRoot).Inject(
		Params{
			api.ID:  h.appId,
			api.Key: key,
		})
	err = h.client.Put(path, value)
	return
}

//
// Delete a fact.
func (h *AppFacts) Delete(name string) (err error) {
	key := api.FactKey(name)
	key.Qualify(h.source)
	path := Path(api.ApplicationFactRoot).Inject(
		Params{
			api.ID:  h.appId,
			api.Key: key,
		})
	err = h.client.Delete(path)
	return
}

//
// Replace facts.
func (h *AppFacts) Replace(facts api.FactMap) (err error) {
	key := api.FactKey("")
	key.Qualify(h.source)
	path := Path(api.ApplicationFactsRoot).Inject(Params{api.ID: h.appId, api.Key: key})
	err = h.client.Put(path, facts)
	return
}

// Analysis returns the analysis API.
func (h *Application) Analysis(id uint) (a Analysis) {
	a = Analysis{
		client: h.client,
		appId:  id,
	}
	return
}

//
// Analysis API.
type Analysis struct {
	client *Client
	appId  uint
}

//
// Create an analysis report.
func (h *Analysis) Create(r *api.Analysis, encoding string, issues, deps io.Reader) (err error) {
	path := Path(api.AppAnalysesRoot).Inject(Params{api.ID: h.appId})
	b, _ := yaml.Marshal(r)
	err = h.client.FileSend(
		path,
		http.MethodPost,
		[]Field{
			{
				Name:     api.FileField,
				Reader:   bytes.NewReader(b),
				Encoding: mime.MIMEYAML,
			},
			{
				Name:     api.IssueField,
				Encoding: encoding,
				Reader:   issues,
			},
			{
				Name:     api.DepField,
				Encoding: encoding,
				Reader:   deps,
			},
		},
		r)
	return
}
