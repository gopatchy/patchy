package api_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gopatchy/api"
	"github.com/gopatchy/patchyc"
	"github.com/stretchr/testify/require"
)

func TestStreamListHeartbeat(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 0)

	time.Sleep(6 * time.Second)

	select {
	case _, ok := <-stream.Chan():
		if ok {
			require.Fail(t, "unexpected list")
		} else {
			require.Fail(t, "unexpected closure")
		}

	default:
	}

	require.Less(t, time.Since(stream.LastEventReceived()), 6*time.Second)
}

func TestStreamListInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{s1[0].Text, s1[1].Text})
}

func TestStreamListAdd(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 0)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "foo", s2[0].Text)
}

func TestStreamListUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	_, err = patchyc.Update[testType](ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
}

func TestStreamListDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 0)
}

func TestStreamListOpts(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Contains(t, []string{"foo", "bar"}, s1[0].Text)
}

func TestStreamListIgnoreIrrelevant(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created1, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Sorts: []string{"+text"}, Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "bar", s1[0].Text)

	_, err = patchyc.Update[testType](ctx, ta.pyc, created1.ID, &testType{Text: "zig"}, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	select {
	case s2 := <-stream.Chan():
		require.Fail(t, "unexpected update", s2)
	default:
	}

	_, err = patchyc.Update[testType](ctx, ta.pyc, created1.ID, &testType{Text: "aaa"}, nil)
	require.NoError(t, err)

	s3 := stream.Read()
	require.NotNil(t, s3, stream.Error())
	require.Len(t, s3, 1)
	require.Equal(t, "aaa", s3[0].Text)
}

func TestStreamListPrev(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := patchyc.StreamList[testType](ctx, ta.pyc, nil)
	require.NoError(t, err)

	defer stream1.Close()

	s1 := stream1.Read()
	require.NotNil(t, s1, stream1.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	s1[0].Num = 5

	stream2, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Prev: s1})
	require.NoError(t, err)

	defer stream2.Close()

	s2 := stream2.Read()
	require.NotNil(t, s2, stream2.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "foo", s2[0].Text)
	require.EqualValues(t, 5, s2[0].Num)
}

func TestStreamListDiffInitial(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)
}

func TestStreamListDiffCreate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 0)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "foo", s2[0].Text)
}

func TestStreamListDiffUpdate(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)
	require.EqualValues(t, 1, s1[0].Num)

	_, err = patchyc.Update[testType](ctx, ta.pyc, created.ID, &testTypeRequest{Text: patchyc.P("bar")}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
	require.EqualValues(t, 1, s2[0].Num)
}

func TestStreamListDiffReplace(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo", Num: 1})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)
	require.EqualValues(t, 1, s1[0].Num)

	_, err = patchyc.Replace[testType](ctx, ta.pyc, created.ID, &testType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
	require.EqualValues(t, 0, s2[0].Num)
}

func TestStreamListDiffDelete(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	err = patchyc.Delete[testType](ctx, ta.pyc, created.ID, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 0)
}

func TestStreamListDiffSort(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{
		Stream: "diff",
		Sorts:  []string{"text"},
		Limit:  1,
	})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
}

func TestStreamListDiffIgnoreIrrelevant(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	created1, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff", Sorts: []string{"+text"}, Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "bar", s1[0].Text)

	_, err = patchyc.Update[testType](ctx, ta.pyc, created1.ID, &testType{Text: "zig"}, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	select {
	case s2 := <-stream.Chan():
		require.Fail(t, "unexpected update", s2)
	default:
	}

	_, err = patchyc.Update[testType](ctx, ta.pyc, created1.ID, &testType{Text: "aaa"}, nil)
	require.NoError(t, err)

	s3 := stream.Read()
	require.NotNil(t, s3, stream.Error())
	require.Len(t, s3, 1)
	require.Equal(t, "aaa", s3[0].Text)
}

func TestStreamListDiffPrev(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream1.Close()

	s1 := stream1.Read()
	require.NotNil(t, s1, stream1.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	s1[0].Num = 5

	stream2, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff", Prev: s1})
	require.NoError(t, err)

	defer stream2.Close()

	s3 := stream2.Read()
	require.NotNil(t, s3, stream2.Error())
	require.Len(t, s3, 1)
	require.Equal(t, "foo", s3[0].Text)
	require.EqualValues(t, 5, s3[0].Num)

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	s4 := stream2.Read()
	require.NotNil(t, s4, stream2.Error())
	require.Len(t, s4, 2)
}

func TestStreamListDiffPrevMiss(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ctx := context.Background()

	_, err := patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff"})
	require.NoError(t, err)

	defer stream1.Close()

	s1 := stream1.Read()
	require.NotNil(t, s1, stream1.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	s1[0].Num = 5

	_, err = patchyc.Create[testType](ctx, ta.pyc, &testType{Text: "bar"})
	require.NoError(t, err)

	stream2, err := patchyc.StreamList[testType](ctx, ta.pyc, &patchyc.ListOpts{Stream: "diff", Prev: s1})
	require.NoError(t, err)

	defer stream2.Close()

	s4 := stream2.Read()
	require.NotNil(t, s4, stream2.Error())
	require.Len(t, s4, 2)
}

func TestStreamListForceDiff(t *testing.T) {
	t.Parallel()

	ta := newTestAPI(t)
	defer ta.shutdown(t)

	ta.api.SetRequestHook(func(r *http.Request, _ *api.API) (*http.Request, error) {
		r.Form.Set("_stream", "diff")
		return r, nil
	})

	resp, err := ta.r().
		SetDoNotParseResponse(true).
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "full").
		Get("testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "text/event-stream", resp.Header().Get("Content-Type"))
	require.Equal(t, "diff", resp.Header().Get("Stream-Format"))
	resp.RawBody().Close()
}
