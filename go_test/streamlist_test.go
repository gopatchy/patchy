package gotest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"test/goclient"
)

func TestStreamListHeartbeat(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	stream, err := c.StreamListTestType(ctx, nil)
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

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 2)
	require.ElementsMatch(t, []string{"foo", "bar"}, []string{s1[0].Text, s1[1].Text})
}

func TestStreamListAdd(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	stream, err := c.StreamListTestType(ctx, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 0)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "foo", s2[0].Text)
}

func TestStreamListUpdate(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	_, err = c.UpdateTestType(ctx, created.ID, &goclient.TestType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
}

func TestStreamListDelete(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	err = c.DeleteTestType(ctx, created.ID, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 0)
}

func TestStreamListOpts(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Contains(t, []string{"foo", "bar"}, s1[0].Text)
}

func TestStreamListIgnoreIrrelevant(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created1, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Sorts: []string{"+text"}, Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "bar", s1[0].Text)

	_, err = c.UpdateTestType(ctx, created1.ID, &goclient.TestType{Text: "zig"}, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	select {
	case s2 := <-stream.Chan():
		require.Fail(t, "unexpected update", s2)
	default:
	}

	_, err = c.UpdateTestType(ctx, created1.ID, &goclient.TestType{Text: "aaa"}, nil)
	require.NoError(t, err)

	s3 := stream.Read()
	require.NotNil(t, s3, stream.Error())
	require.Len(t, s3, 1)
	require.Equal(t, "aaa", s3[0].Text)
}

func TestStreamListPrev(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := c.StreamListTestType(ctx, nil)
	require.NoError(t, err)

	defer stream1.Close()

	s1 := stream1.Read()
	require.NotNil(t, s1, stream1.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	s1[0].Num = 5

	stream2, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Prev: s1})
	require.NoError(t, err)

	defer stream2.Close()

	s2 := stream2.Read()
	require.NotNil(t, s2, stream2.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "foo", s2[0].Text)
	require.EqualValues(t, 5, s2[0].Num)
}

func TestStreamListReconnect(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	closeAllConns(t)

	_, err = c.UpdateTestType(ctx, created.ID, &goclient.TestType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
}

func TestStreamListDiffInitial(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)
}

func TestStreamListDiffCreate(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 0)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "foo", s2[0].Text)
}

func TestStreamListDiffUpdate(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo", Num: 1})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)
	require.EqualValues(t, 1, s1[0].Num)

	_, err = c.UpdateTestType(ctx, created.ID, &goclient.TestType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
	require.EqualValues(t, 1, s2[0].Num)
}

func TestStreamListDiffReplace(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo", Num: 1})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)
	require.EqualValues(t, 1, s1[0].Num)

	_, err = c.ReplaceTestType(ctx, created.ID, &goclient.TestType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
	require.EqualValues(t, 0, s2[0].Num)
}

func TestStreamListDiffDelete(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	err = c.DeleteTestType(ctx, created.ID, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 0)
}

func TestStreamListDiffSort(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{
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

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
}

func TestStreamListDiffIgnoreIrrelevant(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created1, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff", Sorts: []string{"+text"}, Limit: 1})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "bar", s1[0].Text)

	_, err = c.UpdateTestType(ctx, created1.ID, &goclient.TestType{Text: "zig"}, nil)
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	select {
	case s2 := <-stream.Chan():
		require.Fail(t, "unexpected update", s2)
	default:
	}

	_, err = c.UpdateTestType(ctx, created1.ID, &goclient.TestType{Text: "aaa"}, nil)
	require.NoError(t, err)

	s3 := stream.Read()
	require.NotNil(t, s3, stream.Error())
	require.Len(t, s3, 1)
	require.Equal(t, "aaa", s3[0].Text)
}

func TestStreamListDiffPrev(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream1.Close()

	s1 := stream1.Read()
	require.NotNil(t, s1, stream1.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	s1[0].Num = 5

	stream2, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff", Prev: s1})
	require.NoError(t, err)

	defer stream2.Close()

	s3 := stream2.Read()
	require.NotNil(t, s3, stream2.Error())
	require.Len(t, s3, 1)
	require.Equal(t, "foo", s3[0].Text)
	require.EqualValues(t, 5, s3[0].Num)

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	s4 := stream2.Read()
	require.NotNil(t, s4, stream2.Error())
	require.Len(t, s4, 2)
}

func TestStreamListDiffPrevMiss(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	_, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream1, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream1.Close()

	s1 := stream1.Read()
	require.NotNil(t, s1, stream1.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	s1[0].Num = 5

	_, err = c.CreateTestType(ctx, &goclient.TestType{Text: "bar"})
	require.NoError(t, err)

	stream2, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff", Prev: s1})
	require.NoError(t, err)

	defer stream2.Close()

	s4 := stream2.Read()
	require.NotNil(t, s4, stream2.Error())
	require.Len(t, s4, 2)
}

func TestStreamListDiffReconnect(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamListTestType(ctx, &goclient.ListOpts[goclient.TestType]{Stream: "diff"})
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.Len(t, s1, 1)
	require.Equal(t, "foo", s1[0].Text)

	closeAllConns(t)

	_, err = c.UpdateTestType(ctx, created.ID, &goclient.TestType{Text: "bar"}, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.NotNil(t, s2, stream.Error())
	require.Len(t, s2, 1)
	require.Equal(t, "bar", s2[0].Text)
}

func TestStreamListForceDiff(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()

	resp, err := getResty(t).
		SetDoNotParseResponse(true).
		SetHeader("Force-Stream", "diff").
		SetHeader("Accept", "text/event-stream").
		SetQueryParam("_stream", "full").
		Get("api/testtype")
	require.NoError(t, err)
	require.False(t, resp.IsError())
	require.Equal(t, "text/event-stream", resp.Header().Get("Content-Type"))
	require.Equal(t, "diff", resp.Header().Get("Stream-Format"))
}
