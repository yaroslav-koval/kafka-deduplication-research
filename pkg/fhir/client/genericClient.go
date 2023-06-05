package client

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/cerror"
	pkgFHIR "kafka-polygon/pkg/fhir"
	"net/url"
	"strconv"
	"strings"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/edenlabllc/go-fhir-adapter/model/fhir/consts"
)

type GenericClientGetResourceByID interface {
	GetResourceByID(ctx context.Context, resName string, id fhir.ID, dst fhir.AnyResource) error
}

type GenericClientSearchResourceByParams interface {
	SearchResourceByParams(ctx context.Context, resName string, params []*QParam) (*fhir.Bundle, error)
}

type GenericClientPutResource interface {
	PutResource(ctx context.Context, body pkgFHIR.Resource, dst interface{}) error
}

type SearchResult[T pkgFHIR.Resource] struct {
	PagingOptions *SearchResultPagingOptions
	Data          []T
}

type SearchResultPagingOptions struct {
	After AfterOption
	Sort  SortParamsOption
}

const _defaultChunkSize = 100

// GetResourceByID is a generic version of client.GetResourceByID.
// It doesn't require resource name and returns a concrete resource.
// Example: t, err := GetResourceByID[*fhir.Task](ctx, c, id)
func GetResourceByID[T pkgFHIR.Resource](ctx context.Context, c GenericClientGetResourceByID, id fhir.ID) (T, error) {
	dst := new(T)

	if err := c.GetResourceByID(ctx, (*dst).ResourceName(), id, dst); err != nil {
		return *dst, err
	}

	return *dst, nil
}

// SearchResourceByParams is a generic version of client.SearchResourceByParams.
// It doesn't require resource name and returns a slice of concrete resources
// by iterating over received from FHIR bundle entries and adding to the slice
// only those who have fullURL field that contains a name of the given in generic
// parameter resource.
// If you need to extract from bundle resources with fullURL that differs from
// searched resource name (for example you used _revinclude), use the non-generic version.
// Example: t, err := SearchResourceByParams[*fhir.Task](ctx, c, params)
//
// Functional options can be passed: chunkSize(count), limit, after, sorts
// Example t, err := SearchResourceByParams[*fhir.Task](ctx, c, params, WithChunkSize(10), WithLimit(10),
// WithAfter("9c52e89e-41f4-4c51-92c7-e4d677e806eb"),
// WithSortParams([]*SortParam{{Column: "inserted_at", Order: Desc}, {Column: "id", Order: Asc}}))
//
//nolint:gocyclo
func SearchResourceByParams[T pkgFHIR.Resource](
	ctx context.Context,
	c GenericClientSearchResourceByParams,
	params []*QParam,
	sos ...SearchOption) (*SearchResult[T], error) {
	opts := &searchOptions{
		chunkSize: _defaultChunkSize,
	}

	for _, so := range sos {
		so.apply(opts)
	}

	resName := (*new(T)).ResourceName() //nolint:gocritic

	if len(opts.sort) != 0 {
		params = append(params, &QParam{Key: "_sort", Value: opts.sort.String()})
	}

	res := []T{}

	for {
		if opts.limit != nil {
			if len(res)+int(opts.chunkSize) > int(*opts.limit) {
				opts.chunkSize = WithChunkSize(int(*opts.limit) - len(res))
			}
		}

		pagingParams := []*QParam{{Key: "_count", Value: strconv.Itoa(int(opts.chunkSize))}}

		if opts.after != "" {
			pagingParams = append(pagingParams, &QParam{Key: "_after", Value: string(opts.after)})
		}

		b, err := c.SearchResourceByParams(ctx, resName, append(params, pagingParams...))
		if err != nil {
			return nil, err
		}

		for i, e := range b.Entry {
			if strings.Contains(e.FullURL, fmt.Sprintf("/%s/", resName)) {
				v := new(T)

				if err = json.Unmarshal(e.Resource, v); err != nil {
					errKey := pkgFHIR.GenFieldPath(pkgFHIR.PathBundleEntry, i, consts.FieldNameResource)
					errValue := fmt.Sprintf("%s. %s", resName, err.Error())

					return nil, cerror.NewValidationError(ctx, map[string]string{errKey: errValue}).LogError()
				}

				res = append(res, *v)
			}
		}

		q, err := getNextLinkQueryParams(ctx, b)
		if err != nil {
			return nil, err
		}

		opts.after = WithAfter(q.Get("_after"))

		if len(opts.sort) == 0 {
			sortParam := q.Get("_sort")
			if sortParam != "" {
				opts.sort = SortParamsOptionFromString(sortParam)

				params = append(params, &QParam{Key: "_sort", Value: opts.sort.String()})
			}
		}

		if opts.after == "" {
			break
		}

		if opts.limit != nil && len(res) >= int(*opts.limit) {
			break
		}
	}

	return &SearchResult[T]{
		PagingOptions: &SearchResultPagingOptions{
			After: opts.after,
			Sort:  opts.sort,
		},
		Data: res,
	}, nil
}

func getNextLinkQueryParams(ctx context.Context, b *fhir.Bundle) (url.Values, error) {
	for i := range b.Link {
		if b.Link[i].Relation != "next" {
			continue
		}

		u, err := url.Parse(b.Link[i].URL)
		if err != nil {
			return nil, cerror.NewF(ctx, cerror.KindInternal, err.Error()).LogError()
		}

		return u.Query(), nil
	}

	return make(url.Values), nil
}

// PutResource is a generic version of client.PutResource.
// It doesn't require resource name/id and returns a concrete resource.
// Example: t, err := PutResource[*fhir.Parameters](ctx, c, body)
func PutResource[T pkgFHIR.Resource](ctx context.Context, c GenericClientPutResource, body T) (T, error) {
	dst := new(T)

	if err := c.PutResource(ctx, body, dst); err != nil {
		return *dst, err
	}

	return *dst, nil
}
