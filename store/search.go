package store

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/uptrace/bun"
)

// SearchKind says what a hit is.
type SearchKind string

const (
	// KindList is a List, matched on its name.
	KindList SearchKind = "list"
	// KindItem is an Item, matched on its label and quantity.
	KindItem SearchKind = "item"
	// KindNote is a Note block. Notes land in M5; the index is ready for them.
	KindNote SearchKind = "note"
)

// SearchHit is one match, before permissions are applied.
//
// The store deliberately does not filter by who may see what: that decision belongs to
// accessTo in the API layer, and having one place decide it is what stops search
// becoming a way around permissions.
type SearchHit struct {
	Kind SearchKind
	// UID identifies the List or Item the hit belongs to.
	UID string
	// ListID is the List the hit sits on — its own, for a List.
	ListID int64
	// Text is what was indexed, for showing the match.
	Text string
}

// SearchLimit is how many hits a query returns. ⌘K shows a shortlist, not a report.
const SearchLimit = 50

// IndexEntry is one thing to put in the index.
type IndexEntry struct {
	Kind   SearchKind
	UID    string
	ListID int64
	Text   string
}

// Index adds or replaces an entry. Called by the store's own writes, so nothing can be
// created without being findable.
func (s *sqlStore) Index(ctx context.Context, entry IndexEntry) error {
	if err := s.Unindex(ctx, entry.Kind, entry.UID); err != nil {
		return err
	}
	row := &searchIndexModel{
		Kind: string(entry.Kind), UID: entry.UID, ListID: entry.ListID, Text: entry.Text,
	}
	if _, err := s.db.NewInsert().Model(row).Exec(ctx); err != nil {
		return fmt.Errorf("index %s %s: %w", entry.Kind, entry.UID, err)
	}
	return nil
}

// Unindex removes an entry. Removing something absent is not an error.
func (s *sqlStore) Unindex(ctx context.Context, kind SearchKind, uid string) error {
	_, err := s.db.NewDelete().
		Model((*searchIndexModel)(nil)).
		Where("kind = ? AND uid = ?", string(kind), uid).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("unindex %s %s: %w", kind, uid, err)
	}
	return nil
}

// Search returns hits for a query, most relevant first.
//
// The query is what a Member typed into ⌘K. The last word is treated as a prefix, so
// results narrow while they are still typing.
func (s *sqlStore) Search(ctx context.Context, query string) ([]SearchHit, error) {
	terms := searchTerms(query)
	if len(terms) == 0 {
		return nil, nil
	}

	var rows []searchIndexModel
	var err error
	if s.name == "sqlite" {
		err = s.db.NewSelect().
			Model(&rows).
			Where("search_index MATCH ?", sqliteMatch(terms)).
			OrderExpr("bm25(search_index)").
			Limit(SearchLimit).
			Scan(ctx)
	} else {
		err = s.db.NewSelect().
			Model(&rows).
			Where("search @@ to_tsquery('simple', ?)", postgresQuery(terms)).
			OrderExpr("ts_rank(search, to_tsquery('simple', ?)) DESC", postgresQuery(terms)).
			Limit(SearchLimit).
			Scan(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	hits := make([]SearchHit, 0, len(rows))
	for _, row := range rows {
		hits = append(hits, SearchHit{
			Kind: SearchKind(row.Kind), UID: row.UID, ListID: row.ListID, Text: row.Text,
		})
	}
	return hits, nil
}

// searchTerms breaks what a Member typed into words, dropping anything that would
// confuse either query language.
//
// Both FTS5 and to_tsquery have operators of their own — quotes, asterisks, ampersands,
// colons — and a Member typing an apostrophe is not asking for them. Keeping only
// letters and digits means a search box behaves like a search box.
func searchTerms(query string) []string {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			terms = append(terms, strings.ToLower(field))
		}
	}
	return terms
}

// sqliteMatch renders terms for FTS5: every word required, the last one a prefix.
func sqliteMatch(terms []string) string {
	parts := make([]string, len(terms))
	for i, term := range terms {
		if i == len(terms)-1 {
			parts[i] = term + "*"
			continue
		}
		parts[i] = term
	}
	return strings.Join(parts, " AND ")
}

// postgresQuery renders terms for to_tsquery: every word required, the last a prefix.
func postgresQuery(terms []string) string {
	parts := make([]string, len(terms))
	for i, term := range terms {
		if i == len(terms)-1 {
			parts[i] = term + ":*"
			continue
		}
		parts[i] = term
	}
	return strings.Join(parts, " & ")
}

// searchIndexModel is one indexed row.
type searchIndexModel struct {
	bun.BaseModel `bun:"table:search_index,alias:search_index"`

	Kind   string `bun:"kind,notnull"`
	UID    string `bun:"uid,notnull"`
	ListID int64  `bun:"list_id,notnull"`
	Text   string `bun:"text,notnull"`
}
