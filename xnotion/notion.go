package xnotion

import (
	"context"

	"github.com/jomei/notionapi"
	"github.com/samber/oops"
)

type Client struct {
	API *notionapi.Client
}

func NewClient(token string) *Client {
	return &Client{
		API: notionapi.NewClient(notionapi.Token(token)),
	}
}

// DatabasePages fetches all pages from a Notion database.
func (c *Client) DatabasePages(ctx context.Context, id notionapi.DatabaseID) ([]notionapi.Page, error) {
	var pages = make([]notionapi.Page, 0)
	var cursor notionapi.Cursor
	for {
		res, err := c.API.Database.Query(ctx, id, &notionapi.DatabaseQueryRequest{
			StartCursor: cursor,
		})
		if err != nil {
			return nil, oops.With("db_id", id).Wrap(err)
		}
		pages = append(pages, res.Results...)
		if !res.HasMore {
			break
		}
		cursor = res.NextCursor
	}
	return pages, nil
}
