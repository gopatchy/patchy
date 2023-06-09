package gotest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"test/goclient"
)

func TestDeleteSuccess(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	get, err := c.GetTestType(ctx, created.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, get)
	require.Equal(t, "foo", get.Text)

	err = c.DeleteTestType(ctx, created.ID, nil)
	require.NoError(t, err)

	get, err = c.GetTestType(ctx, created.ID, nil)
	require.NoError(t, err)
	require.Nil(t, get)
}

func TestDeleteInvalidID(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	err := c.DeleteTestType(ctx, "doesnotexist", nil)
	require.Error(t, err)
}

func TestDeleteTwice(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	err = c.DeleteTestType(ctx, created.ID, nil)
	require.NoError(t, err)

	err = c.DeleteTestType(ctx, created.ID, nil)
	require.Error(t, err)
}

func TestDeleteStream(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	stream, err := c.StreamGetTestType(ctx, created.ID, nil)
	require.NoError(t, err)

	defer stream.Close()

	s1 := stream.Read()
	require.NotNil(t, s1, stream.Error())
	require.NoError(t, stream.Error())
	require.Equal(t, "foo", s1.Text)

	err = c.DeleteTestType(ctx, created.ID, nil)
	require.NoError(t, err)

	s2 := stream.Read()
	require.Nil(t, s2, stream.Error())
	require.Error(t, stream.Error())
}

func TestDeleteIfMatchETagSuccess(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	err = c.DeleteTestType(ctx, created.ID, &goclient.UpdateOpts[goclient.TestType]{Prev: created})
	require.NoError(t, err)

	get, err := c.GetTestType(ctx, created.ID, nil)
	require.NoError(t, err)
	require.Nil(t, get)
}

func TestDeleteIfMatchETagMismatch(t *testing.T) {
	t.Parallel()

	defer registerTest(t)()
	c := getClient(t)
	ctx := context.Background()

	created, err := c.CreateTestType(ctx, &goclient.TestType{Text: "foo"})
	require.NoError(t, err)

	_, err = c.UpdateTestType(ctx, created.ID, &goclient.TestType{Text: "bar"}, nil)
	require.NoError(t, err)

	err = c.DeleteTestType(ctx, created.ID, &goclient.UpdateOpts[goclient.TestType]{Prev: created})
	require.Error(t, err)

	get, err := c.GetTestType(ctx, created.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, get)
	require.Equal(t, "bar", get.Text)
}
