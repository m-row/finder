package finder

import "math"

// metaBuilder builds the meta object of request containing pagination
// column info and total of result model.
func metaBuilder(m *Meta, urlQuery *URLQuery) {
	var offset, fraction uint64

	m.FirstPage = 1

	// var page, paginate uint64
	if urlQuery.Paginate == 0 {
		m.Paginate = 12
	} else {
		m.Paginate = urlQuery.Paginate
	}
	if urlQuery.Page <= 1 {
		m.CurrentPage = 1
		offset = 0
	} else {
		m.CurrentPage = urlQuery.Page
		offset = m.CurrentPage*m.Paginate - m.Paginate
	}

	m.LastPage = uint64(
		math.Ceil(
			float64(m.Total) / float64(m.Paginate),
		),
	)
	remainingItems := m.Total % m.Paginate
	if remainingItems == 0 {
		fraction = m.Paginate
	} else {
		fraction = remainingItems
	}

	if offset == 0 {
		m.From = 1
	} else {
		m.From = offset + 1
	}

	if m.CurrentPage == m.LastPage {
		m.To = offset + fraction
	} else {
		m.To = offset + m.Paginate
	}
}
