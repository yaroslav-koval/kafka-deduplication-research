package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"kafka-polygon/pkg/fhir/client"
	"net/url"
	"strconv"
	"strings"

	"github.com/edenlabllc/go-fhir-adapter/model/fhir"
	"github.com/valyala/fasthttp"
)

func (s *clientTestSuite) TestGetResourceByIDGeneric() {
	ctx := context.Background()
	resName := "Task"
	resID := fhir.ID("123")
	expTask := &fhir.Task{ID: string(resID)}
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		// http method and headers are not affected by generic method and covered in TestGetResourceByID
		s.Equal(fmt.Sprintf("%s/%s/%s", s.fhirClientCfg.HostSearch, resName, resID), req.URI().String())

		b, _ := json.Marshal(expTask)

		resp.SetBody(b)
		resp.SetStatusCode(fasthttp.StatusOK)

		return nil
	}

	actTask, err := client.GetResourceByID[*fhir.Task](ctx, s.fhirClient, resID)
	s.NoError(err)
	s.Equal(expTask, actTask)
}

func (s *clientTestSuite) TestSearchResourceByParamsGeneric() {
	ctx := context.Background()
	resName := "Task"

	expTasks := []*fhir.Task{{ID: "1"}, {ID: "2"}}

	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		// http method, headers and query params are not affeced by generic method
		// and covered in TestSearchResourceByParams
		s.True(strings.HasPrefix(req.URI().String(), fmt.Sprintf("%s/%s", s.fhirClientCfg.HostSearch, resName)))
		resp.SetStatusCode(fasthttp.StatusOK)

		t1Bytes, _ := json.Marshal(expTasks[0])
		t2Bytes, _ := json.Marshal(expTasks[1])

		b := &fhir.Bundle{
			Entry: []*fhir.BundleEntry{
				{FullURL: "some other resource", Resource: t1Bytes},
				{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t1Bytes},
				{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/2", Resource: t2Bytes},
			},
		}

		bBytes, _ := json.Marshal(b)

		_, _ = resp.BodyWriter().Write(bBytes)

		return nil
	}

	actTasksRes, err := client.SearchResourceByParams[*fhir.Task](ctx, s.fhirClient, nil)
	s.NoError(err)
	s.Equal(expTasks, actTasksRes.Data)
}

func (s *clientTestSuite) TestSearchResourceByParamsGenericPagingOptions() {
	ctx := context.Background()
	resName := "Task"
	expTasks := []*fhir.Task{{ID: "1"}, {ID: "2"}, {ID: "3"}}

	// assert default options
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		u, _ := url.Parse(req.URI().String())
		q := u.Query()
		s.Equal(strconv.Itoa(100), q.Get("_count"))

		t1Bytes, _ := json.Marshal(expTasks[0])

		b := &fhir.Bundle{
			Entry: []*fhir.BundleEntry{
				{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t1Bytes},
			},
		}

		bBytes, _ := json.Marshal(b)

		_, _ = resp.BodyWriter().Write(bBytes)

		return nil
	}

	_, err := client.SearchResourceByParams[*fhir.Task](ctx, s.fhirClient, nil)
	s.NoError(err)

	// _after, _sort is set
	// _after reference working correct
	// must make request exactly 3 times (_after not set in 3 bundle)
	expChunkSize := 1
	expLimit := 100

	pageCounter := 0
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		u, _ := url.Parse(req.URI().String())
		q := u.Query()
		s.Equal(strconv.Itoa(expChunkSize), q.Get("_count"))

		var b *fhir.Bundle

		switch pageCounter {
		case 0:
			t1Bytes, _ := json.Marshal(expTasks[0])

			b = &fhir.Bundle{
				Entry: []*fhir.BundleEntry{
					{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t1Bytes},
				},
				Link: []*fhir.BundleLink{
					{Relation: "next", URL: s.fhirClientCfg.HostSearch + "/" + resName + "/1?_after=1&_sort=-inserted_at,id"},
				},
			}

		case 1:
			t2Bytes, _ := json.Marshal(expTasks[1])

			b = &fhir.Bundle{
				Entry: []*fhir.BundleEntry{
					{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t2Bytes},
				},
				Link: []*fhir.BundleLink{
					{Relation: "next", URL: s.fhirClientCfg.HostSearch + "/" + resName + "/1?_after=ref2&_sort=id"},
				},
			}

		case 2:
			t3Bytes, _ := json.Marshal(expTasks[2])

			b = &fhir.Bundle{
				Entry: []*fhir.BundleEntry{
					{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t3Bytes},
				},
			}
		case 3:
			t3Bytes, _ := json.Marshal(expTasks[2])

			b = &fhir.Bundle{
				Entry: []*fhir.BundleEntry{
					{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t3Bytes},
				},
			}
		}

		bBytes, _ := json.Marshal(b)

		_, _ = resp.BodyWriter().Write(bBytes)

		pageCounter++

		return nil
	}

	actTasksRes, err := client.SearchResourceByParams[*fhir.Task](ctx, s.fhirClient, nil,
		client.WithChunkSize(expChunkSize), client.WithLimit(expLimit))
	s.NoError(err)
	s.Equal(3, pageCounter)
	s.Equal(expTasks, actTasksRes.Data)
	s.EqualValues("", actTasksRes.PagingOptions.After)
	s.EqualValues("-inserted_at,id", actTasksRes.PagingOptions.Sort.String())

	pageCounter = 0

	// sort
	actTasksRes, err = client.SearchResourceByParams[*fhir.Task](ctx, s.fhirClient, nil,
		client.WithChunkSize(expChunkSize), client.WithLimit(expLimit),
		client.WithSortParams([]*client.SortParam{
			{Column: "column_1", Order: client.Desc},
			{Column: "column_2", Order: client.Asc},
			{Column: "column_3", Order: client.Desc}}))
	s.NoError(err)
	s.EqualValues("-column_1,column_2,-column_3", actTasksRes.PagingOptions.Sort.String())

	pageCounter = 0

	// chunkSize, limit
	s.httpClient.DoFunc = func(req *fasthttp.Request, resp *fasthttp.Response) error {
		// FHIR logic duplication
		u, _ := url.Parse(req.URI().String())
		q := u.Query()
		countStr := q.Get("_count")
		count, _ := strconv.Atoi(countStr)

		afterStr := q.Get("_after")

		be := []*fhir.BundleEntry{}
		b := &fhir.Bundle{}

		var i int
		if afterStr == "" {
			for i = 0; i < count; i++ {
				t, _ := json.Marshal(expTasks[i])

				be = append(be, &fhir.BundleEntry{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t})
			}
		} else {
			after, _ := strconv.Atoi(afterStr)
			last := after + count
			for i = after; i < last; i++ {
				t, _ := json.Marshal(expTasks[i])

				be = append(be, &fhir.BundleEntry{FullURL: s.fhirClientCfg.HostSearch + "/" + resName + "/1", Resource: t})
			}
		}

		b.Link = []*fhir.BundleLink{
			{Relation: "next", URL: s.fhirClientCfg.HostSearch + "/" + resName + "/1?_count=" + countStr + "&_after=" + strconv.Itoa(i)},
		}

		b.Entry = be
		bBytes, _ := json.Marshal(b)

		_, _ = resp.BodyWriter().Write(bBytes)

		pageCounter++

		return nil
	}

	expTasks = []*fhir.Task{}
	expChunkSize = 3
	expLimit = 8

	for i := 0; i < 15; i++ {
		expTasks = append(expTasks, &fhir.Task{ID: strconv.Itoa(i)})
	}

	actTasksRes, err = client.SearchResourceByParams[*fhir.Task](ctx, s.fhirClient, nil,
		client.WithChunkSize(expChunkSize), client.WithLimit(expLimit))
	s.NoError(err)
	s.Equal(3, pageCounter)
	s.Equal(expLimit, len(actTasksRes.Data))
	s.Equal(expTasks[:expLimit], actTasksRes.Data)

	pageCounter = 0
}
