package xnotion

import (
	"context"
	"strings"
	"time"

	"github.com/jomei/notionapi"
)

func (c *Client) ParsePropertyString(ps notionapi.Properties, key string) (string, bool) {
	p, ok := ps[key]
	if !ok {
		return "", false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeTitle:
		s := p.(*notionapi.TitleProperty).Title
		resp := ""
		for _, v := range s {
			resp += v.PlainText
		}
		return strings.TrimSpace(resp), true
	case notionapi.PropertyTypeRichText:
		s := p.(*notionapi.RichTextProperty).RichText
		resp := ""
		for _, v := range s {
			resp += v.PlainText
		}
		return strings.TrimSpace(resp), true
	case notionapi.PropertyTypeText:
		s := p.(*notionapi.TextProperty).Text
		resp := ""
		for _, v := range s {
			resp += v.PlainText
		}
		return strings.TrimSpace(resp), true
	case notionapi.PropertyTypeSelect:
		return p.(*notionapi.SelectProperty).Select.Name, true
	case notionapi.PropertyTypeStatus:
		return p.(*notionapi.StatusProperty).Status.Name, true
	case notionapi.PropertyTypeURL:
		return p.(*notionapi.URLProperty).URL, true
	case notionapi.PropertyTypeEmail:
		return p.(*notionapi.EmailProperty).Email, true
	case notionapi.PropertyTypePhoneNumber:
		return p.(*notionapi.PhoneNumberProperty).PhoneNumber, true
	case notionapi.PropertyTypeDate:
		dt := p.(*notionapi.DateProperty).Date
		if dt != nil && dt.Start != nil {
			return time.Time(*dt.Start).Format("2006-01-02"), true
		}
		return "", true
	case notionapi.PropertyTypeRelation:
		r := p.(*notionapi.RelationProperty).Relation
		if len(r) == 0 {
			return "", true
		}
		// get the page, and parse the ID field
		rp, err := c.API.Page.Get(context.Background(), r[0].ID)
		if err != nil {
			return "", false
		}
		v, ok := c.ParsePropertyString(rp.Properties, "ID")
		if !ok {
			return "", false
		}
		return v, true
	default:
		return "", false
	}
}

func (c *Client) ParsePropertyStringArrayByComma(ps notionapi.Properties, key string) ([]string, bool) {
	row, ok := c.ParsePropertyString(ps, key)
	if !ok {
		return nil, false
	}
	return strings.Split(row, ","), true
}

func (c *Client) ParsePropertyStringArrayByNewline(ps notionapi.Properties, key string) ([]string, bool) {
	var resp = make([]string, 0)
	p, ok := ps[key]
	if !ok {
		return nil, false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeTitle:
		s := p.(*notionapi.TitleProperty).Title
		for _, v := range s {
			if strings.TrimSpace(v.PlainText) != "" {
				resp = append(resp, v.PlainText)
			}
		}
	case notionapi.PropertyTypeRichText:
		s := p.(*notionapi.RichTextProperty).RichText
		for _, v := range s {
			if strings.TrimSpace(v.PlainText) != "" {
				resp = append(resp, v.PlainText)
			}
		}
	case notionapi.PropertyTypeText:
		s := p.(*notionapi.TextProperty).Text
		for _, v := range s {
			if strings.TrimSpace(v.PlainText) != "" {
				resp = append(resp, v.PlainText)
			}
		}
	case notionapi.PropertyTypeRelation:
		r := p.(*notionapi.RelationProperty).Relation
		if len(r) == 0 {
			return resp, true
		}
		for _, v := range r {
			// get the page, and parse the ID field
			rp, err := c.API.Page.Get(context.Background(), v.ID)
			if err != nil {
				return nil, false
			}
			v, ok := c.ParsePropertyString(rp.Properties, "ID")
			if !ok {
				return nil, false
			}
			resp = append(resp, v)
		}
	default:
		return nil, false
	}
	return resp, true
}

func (c *Client) ParsePropertyTime(ps notionapi.Properties, key string) (*time.Time, bool, error) {
	p, ok := ps[key]
	if !ok {
		return nil, false, nil
	}
	switch p.GetType() {
	case notionapi.PropertyTypeRichText:
		s := p.(*notionapi.RichTextProperty).RichText
		if len(s) == 0 {
			return nil, true, nil
		}
		str := p.(*notionapi.RichTextProperty).RichText[0].PlainText
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return nil, false, err
		}
		return &t, true, nil
	case notionapi.PropertyTypeDate:
		t := time.Time(*p.(*notionapi.DateProperty).Date.Start)
		return &t, true, nil
	default:
		return nil, false, nil
	}
}

func (c *Client) ParsePropertyFloat(ps notionapi.Properties, key string) (float64, bool) {
	p, ok := ps[key]
	if !ok {
		return 0, false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeNumber:
		return p.(*notionapi.NumberProperty).Number, true
	default:
		return 0, false
	}
}

func (c *Client) ParsePropertyInt(ps notionapi.Properties, key string) (int, bool) {
	p, ok := ps[key]
	if !ok {
		return 0, false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeNumber:
		return int(p.(*notionapi.NumberProperty).Number), true
	default:
		return 0, false
	}
}

func (c *Client) ParsePropertyInt64(ps notionapi.Properties, key string) (int64, bool) {
	p, ok := ps[key]
	if !ok {
		return 0, false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeNumber:
		return int64(p.(*notionapi.NumberProperty).Number), true
	default:
		return 0, false
	}
}

func (c *Client) ParsePropertyBoolPointer(ps notionapi.Properties, key string) (*bool, bool) {
	p, ok := ps[key]
	if !ok {
		return nil, false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeCheckbox:
		return &p.(*notionapi.CheckboxProperty).Checkbox, true
	default:
		return nil, false
	}
}

func (c *Client) ParsePropertyBool(ps notionapi.Properties, key string) (bool, bool) {
	p, ok := ps[key]
	if !ok {
		return false, false
	}
	switch p.GetType() {
	case notionapi.PropertyTypeCheckbox:
		return p.(*notionapi.CheckboxProperty).Checkbox, true
	default:
		return false, false
	}
}
