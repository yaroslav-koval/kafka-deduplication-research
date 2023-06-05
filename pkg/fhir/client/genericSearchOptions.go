package client

import "strings"

type SearchOption interface {
	apply(options *searchOptions)
}

type searchOptions struct {
	chunkSize ChunkSizeOption
	limit     *LimitOption
	after     AfterOption
	sort      SortParamsOption
}

type ChunkSizeOption int

func (cs ChunkSizeOption) apply(options *searchOptions) {
	options.chunkSize = cs
}

func WithChunkSize(size int) ChunkSizeOption {
	return ChunkSizeOption(size)
}

type LimitOption int

func (lo LimitOption) apply(options *searchOptions) {
	options.limit = &lo
}

func WithLimit(limit int) LimitOption {
	return LimitOption(limit)
}

type AfterOption string

func (ao AfterOption) apply(options *searchOptions) {
	options.after = ao
}

func WithAfter(s string) AfterOption {
	return AfterOption(s)
}

type SortParamsOption []*SortParam

type SortParam struct {
	Column string
	Order  Order
}

type Order int

const (
	Asc Order = iota
	Desc
)

func (spo SortParamsOption) apply(options *searchOptions) {
	options.sort = spo
}

func (spo SortParamsOption) String() string {
	if len(spo) == 0 {
		return ""
	}

	sb := strings.Builder{}

	for _, sp := range spo {
		if sp.Order == Desc {
			sb.WriteByte('-')
		}

		sb.WriteString(sp.Column)
		sb.WriteByte(',')
	}

	res := []byte(sb.String())

	return string(res[:len(res)-1])
}

func WithSortParams(sp []*SortParam) SortParamsOption {
	return sp
}

func SortParamsOptionFromString(columns string) SortParamsOption {
	if columns == "" {
		return nil
	}

	res := []*SortParam{}

	for _, c := range strings.Split(columns, ",") {
		if strings.HasPrefix(c, "-") {
			res = append(res, &SortParam{
				Column: c[1:],
				Order:  Desc,
			})

			continue
		}

		res = append(res, &SortParam{
			Column: c,
			Order:  Asc,
		})
	}

	return res
}
