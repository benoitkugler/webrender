package tree

import (
	"fmt"
	"strings"

	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"

	pr "github.com/benoitkugler/webrender/css/properties"
)

// Handle target-counter, target-counters && target-text.
//
// The TargetCollector is a structure providing required targets"
// counterValues && stuff needed to build Pending targets later,
// when the layout of all targetted anchors has been done.
type RemakeState struct {
	// first occurrence of anchor
	Anchors []string
	// first occurrence of content-CounterLookupItem
	ContentLookups              []*CounterLookupItem
	ContentChanged, PagesWanted bool
}

// ResumeStack is a tree of index boxes where
// layout has been interrupted and should resume,
// following the box tree structure.
// At each level, several boxes may be selected.
type ResumeStack map[int]ResumeStack

// Unpack returns the first element and the corresponding stack.
// It will panic if the stack is empty.
func (r ResumeStack) Unpack() (int, ResumeStack) {
	for k, v := range r {
		return k, v
	}
	panic("invalid use of Unpack on an empty stack")
}

func (r ResumeStack) String() string {
	if len(r) == 0 {
		return "-"
	}
	return fmt.Sprintf("%v", map[int]ResumeStack(r))
}

// Equals returns true if the two stacks are deeply equals.
// nil and 0-sized map are compared equal.
func (r ResumeStack) Equals(other ResumeStack) bool {
	if len(r) != len(other) {
		return false
	}
	for k, v1 := range r {
		v2, has := other[k]
		if !has || !v1.Equals(v2) {
			return false
		}
	}
	return true
}

type PageState struct {
	QuoteDepth    []int
	CounterValues CounterValues
	CounterScopes []utils.Set
}

// Copy returns a deep copy.
func (s PageState) Copy() PageState {
	out := PageState{}
	out.QuoteDepth = append([]int{}, s.QuoteDepth...)
	out.CounterValues = s.CounterValues.Copy()
	out.CounterScopes = make([]utils.Set, len(s.CounterScopes))
	for i, v := range s.CounterScopes {
		out.CounterScopes[i] = v.Copy()
	}
	return out
}

// Equal returns `true` for deep equality
func (s PageState) Equal(other PageState) bool {
	if len(s.CounterScopes) != len(other.CounterScopes) {
		return false
	}
	for i := range s.CounterScopes {
		if !s.CounterScopes[i].Equal(other.CounterScopes[i]) {
			return false
		}
	}
	return equalInts(s.QuoteDepth, other.QuoteDepth) && s.CounterValues.Equal(other.CounterValues)
}

type PageBreak struct {
	Break string
	Page  pr.Page
}

type PageMaker struct {
	InitialResumeAt  ResumeStack
	InitialPageState PageState
	InitialNextPage  PageBreak
	RemakeState      RemakeState
	RightPage        bool
}

type Box interface {
	CachedCounterValues() CounterValues
	SetCachedCounterValues(cv CounterValues)
	MissingLink() Box
	SetMissingLink(b Box)
	GetBookmarkLabel() string
}

type CounterValues map[string][]int

// Copy performs a deep copy of c
func (c CounterValues) Copy() CounterValues {
	out := make(CounterValues, len(c))
	for k, v := range c {
		out[k] = append([]int{}, v...)
	}
	return out
}

func (c CounterValues) Update(other CounterValues) {
	for k, v := range other {
		c[k] = v
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ai := range a {
		if ai != b[i] {
			return false
		}
	}
	return true
}

// Equal deeply compare each elements of c and other
func (c CounterValues) Equal(other CounterValues) bool {
	if len(c) != len(other) {
		return false
	}
	for k, v := range c {
		if !equalInts(v, other[k]) {
			return false
		}
	}
	return true
}

type functionKey struct {
	SourceBox Box
	CssToken  string
}

func NewFunctionKey(sourceBox Box, cssToken string) functionKey {
	return functionKey{CssToken: cssToken, SourceBox: sourceBox}
}

type funcStore = map[functionKey]ParseFunc

type ParseFunc = func(CounterValues)

type targetState uint8

const (
	pending targetState = iota
	upToDate
	undefined
)

// Item controlling Pending targets and page based target counters.
//
// Collected in the TargetCollector"s “items“.
type TargetLookupItem struct {
	// Required by target-counter and target-counters to access the
	// target's .cachedCounterValues.
	// Needed for target-text via extract_text.
	TargetBox Box

	// Functions that have to been called to check Pending targets.
	// Keys are (sourceBox, cssToken).
	parseAgainFunctions funcStore

	// TargetBox's pageCounters during pagination
	CachedPageCounterValues CounterValues

	state targetState

	// Anchor position during pagination (pageNumber - 1)
	PageMakerIndex int
}

func newTargetLookupItem(state targetState) *TargetLookupItem {
	return &TargetLookupItem{state: state, parseAgainFunctions: funcStore{}, CachedPageCounterValues: CounterValues{}}
}

func (t TargetLookupItem) IsUpToDate() bool { return t.state == upToDate }

type optionnalInt struct {
	int
	none bool
}

func NewOptionnalInt(i int) optionnalInt {
	return optionnalInt{int: i}
}

// Item controlling page based counters.
//
// Collected in the TargetCollector's “CounterLookupItems“.
type CounterLookupItem struct {
	// Function that have to been called to check Pending counter.
	ParseAgain ParseFunc

	// Missing counters and target counters
	MissingCounters       utils.Set
	MissingTargetCounters map[string]utils.Set

	// Targeting box's pageCounters during pagination
	CachedPageCounterValues CounterValues

	// Box position during pagination (pageNumber - 1)
	PageMakerIndex optionnalInt

	// Marker for remakePage
	Pending bool
}

func NewCounterLookupItem(parseAgain ParseFunc, missingCounters utils.Set, missingTargetCounters map[string]utils.Set) *CounterLookupItem {
	return &CounterLookupItem{
		ParseAgain:              parseAgain,
		MissingCounters:         missingCounters,
		MissingTargetCounters:   missingTargetCounters,
		PageMakerIndex:          optionnalInt{none: true},
		CachedPageCounterValues: CounterValues{},
	}
}

// Collector of HTML targets used by CSS content with “target-*“.
type TargetCollector struct {
	// Lookup items for targets and page counters
	TargetLookupItems  map[string]*TargetLookupItem
	CounterLookupItems map[functionKey]*CounterLookupItem

	// When collecting is true, computeContentList() collects missing
	// page counters in CounterLookupItems. Otherwise, it mixes in the
	// TargetLookupItem's CachedPageCounterValues.
	// Is switched to false in CheckPendingTargets().
	collecting bool

	// hadPendingTargets is set to true when a target is needed but has
	// not been seen yet. CheckPendingTargets then uses this information
	// to call the needed ParseAgain functions.
	hadPendingTargets bool
}

func NewTargetCollector() TargetCollector {
	return TargetCollector{
		TargetLookupItems:  map[string]*TargetLookupItem{},
		CounterLookupItems: map[functionKey]*CounterLookupItem{},
		collecting:         true,
	}
}

func (t *TargetCollector) IsCollecting() bool { return t.collecting }

// Get anchor name from string or uri token.
func AnchorNameFromToken(anchorToken pr.ContentProperty) string {
	asString, _ := anchorToken.Content.(pr.String)
	asUrl, _ := anchorToken.Content.(pr.NamedString)
	if anchorToken.Type == "string" && strings.HasPrefix(string(asString), "#") {
		return string(asString[1:])
	} else if anchorToken.Type == "url" && asUrl.Name == "internal" {
		return asUrl.String
	}
	return ""
}

// Create a TargetLookupItem for the given `anchorName`.
func (tc *TargetCollector) collectAnchor(anchorName string) {
	if anchorName != "" {
		if _, has := tc.TargetLookupItems[anchorName]; has {
			logger.WarningLogger.Printf("Anchor defined twice: %s \n", anchorName)
		} else {
			tc.TargetLookupItems[anchorName] = newTargetLookupItem(pending)
		}
	}
}

// Get a TargetLookupItem corresponding to “anchorToken“.
//
// If it is already filled by a previous anchor-Element, the status is
// "up-to-date". Otherwise, it is "Pending", we must parse the whole
// tree again.
func (tc *TargetCollector) LookupTarget(anchorToken pr.ContentProperty, sourceBox Box, cssToken string, parseAgain ParseFunc) *TargetLookupItem {
	anchorName := AnchorNameFromToken(anchorToken)
	item, in := tc.TargetLookupItems[anchorName]
	if !in {
		item = newTargetLookupItem(undefined)
	}

	if item.state == pending {
		tc.hadPendingTargets = true
		key := functionKey{SourceBox: sourceBox, CssToken: cssToken}
		if _, in := item.parseAgainFunctions[key]; !in {
			item.parseAgainFunctions[key] = parseAgain
		}
	}

	if item.state == undefined {
		logger.WarningLogger.Printf("Content discarded: target points to undefined anchor '%s' \n", anchorToken)
	}

	return item
}

// Store a target called “anchorName“.
//
// If there is a Pending TargetLookupItem, it is updated. Only previously
// collected anchors are stored.
func (tc *TargetCollector) StoreTarget(anchorName string, targetCounterValues CounterValues, targetBox Box) {
	item := tc.TargetLookupItems[anchorName]
	if item != nil && item.state == pending {
		item.state = upToDate
		item.TargetBox = targetBox
		// Store the counterValues in the TargetBox like
		// computeContentList does.
		if targetBox.CachedCounterValues() == nil {
			targetBox.SetCachedCounterValues(targetCounterValues.Copy())
		}
	}
}

// Collect missing (probably page-based) counters during formatting.
//
// The “MissingCounters“ are re-used during pagination.
//
// The “missingLink“ attribute added to the parentBox is required to
// connect the paginated boxes to their originating “parentBox“.
func (tc *TargetCollector) CollectMissingCounters(parentBox Box, cssToken string,
	parseAgainFunction ParseFunc, missingCounters utils.Set, missingTargetCounters map[string]utils.Set,
) {
	// No counter collection during pagination
	if !tc.collecting {
		return
	}

	// No need to add empty miss-lists
	if len(missingCounters) > 0 || len(missingTargetCounters) > 0 {
		if parentBox.MissingLink() == nil {
			parentBox.SetMissingLink(parentBox)
		}
		counterLookupItem := NewCounterLookupItem(
			parseAgainFunction, missingCounters,
			missingTargetCounters)
		key := functionKey{SourceBox: parentBox, CssToken: cssToken}
		if _, in := tc.CounterLookupItems[key]; !in {
			tc.CounterLookupItems[key] = counterLookupItem
		}

	}
}

// Check Pending targets if needed.
func (tc *TargetCollector) CheckPendingTargets() {
	if tc.hadPendingTargets {
		for _, item := range tc.TargetLookupItems {
			for _, function := range item.parseAgainFunctions {
				function(nil)
			}
		}
		tc.hadPendingTargets = false
	}
	// Ready for pagination
	tc.collecting = false
}

// Store target's current “PageMakerIndex“ and page counter values.
//
// Eventually update associated targeting boxes.
func (tc *TargetCollector) CacheTargetPageCounters(anchorName string, pageCounterValues CounterValues, pageMakerIndex int,
	pageMaker []PageMaker,
) {
	// Only store page counters when paginating
	if tc.collecting {
		return
	}

	item := tc.TargetLookupItems[anchorName]
	if item != nil && item.IsUpToDate() {
		item.PageMakerIndex = pageMakerIndex
		if !item.CachedPageCounterValues.Equal(pageCounterValues) {
			item.CachedPageCounterValues = pageCounterValues.Copy()
		}
	}

	// Spread the news: update boxes affected by a change in the
	// anchor"s page counter values.
	for key, item := range tc.CounterLookupItems {
		// (_, cssToken) = key
		// Only update items that need counters in their content
		if key.CssToken != "content" {
			continue
		}

		// Don"t update if item has no missing target counter
		missingCounters := item.MissingTargetCounters[anchorName]
		if missingCounters == nil {
			continue
		}

		// Pending marker for remakePage
		if item.PageMakerIndex.none || item.PageMakerIndex.int >= len(pageMaker) {
			item.Pending = true
			continue
		}

		// TODO: Is the item at all interested inthe new
		// pageCounterValues? It probably is and this check is a
		// brake.
		for counterName := range missingCounters {
			if _, in := pageCounterValues[counterName]; in {
				pageMaker[item.PageMakerIndex.int].RemakeState.ContentChanged = true
				item.ParseAgain(item.CachedPageCounterValues)
				break
			}
		}
		// Hint: the box's own cached page counters trigger a
		// separate "contentChanged".
	}
}
