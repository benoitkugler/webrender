package layout

import (
	"fmt"
	"sort"

	pr "github.com/benoitkugler/webrender/css/properties"
	kw "github.com/benoitkugler/webrender/css/properties/keywords"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/html/tree"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
)

// Layout for grid containers and grid-items.

func isLength(sizing pr.DimOrS) bool { return sizing.Unit != 0 && sizing.Unit != pr.Fr }

func isFr(sizing pr.DimOrS) bool { return sizing.Unit == pr.Fr }

func intersect(position1, size1, position2, size2 int) bool {
	return position1 < position2+size2 && position2 < position1+size1
}

func intersectWithChildren(x, y, width, height int, positions map[Box]rect) bool {
	for _, rect := range positions {
		fullX, fullY, fullWidth, fullHeight := rect.unpack()
		xIntersect := intersect(x, width, fullX, fullWidth)
		yIntersect := intersect(y, height, fullY, fullHeight)
		if xIntersect && yIntersect {
			return true
		}
	}
	return false
}

func getTemplateTracks(tracks pr.GridTemplate) []pr.GridSpec {
	if tracks.Tag == pr.None {
		tracks.Names = []pr.GridSpec{pr.GridNames{}}
	}
	if tracks.Tag == pr.Subgrid {
		// TODO: Support subgrids.
		logger.WarningLogger.Println("Subgrids are unsupported")
		return []pr.GridSpec{pr.GridNames{}}
	}
	var tracksList []pr.GridSpec
	for i, track := range tracks.Names {
		if i%2 != 0 {
			// Track size.
			if trackR, isRepeat := track.(pr.GridRepeat); isRepeat {
				repeatNumber := trackR.Repeat
				if repeatNumber == pr.RepeatAutoFill || repeatNumber == pr.RepeatAutoFit {
					// TODO: Respect auto-fit && auto-fill.
					logger.WarningLogger.Println(`"auto-fit" and "auto-fill" are unsupported in repeat()`)
					repeatNumber = 1
				}
				for _c := 0; _c < repeatNumber; _c++ {
					for j, repeatTrack := range trackR.Names {
						if j%2 != 0 {
							// Track size in repeat.
							tracksList = append(tracksList, repeatTrack)
						} else {
							// Line names in repeat.
							if len(tracksList)%2 != 0 {
								tracksList[len(tracksList)-1] = append(tracksList[len(tracksList)-1].(pr.GridNames), repeatTrack.(pr.GridNames)...)
							} else {
								tracksList = append(tracksList, repeatTrack)
							}
						}
					}
				}
			} else {
				tracksList = append(tracksList, track)
			}
		} else {
			// Line names.
			if len(tracksList)%2 != 0 {
				tracksList[len(tracksList)-1] = append(tracksList[len(tracksList)-1].(pr.GridNames), track.(pr.GridNames)...)
			} else {
				tracksList = append(tracksList, track)
			}
		}
	}
	return tracksList
}

type maybeInt struct {
	valid bool
	i     int
}

func getLine(line pr.GridLine, lines []pr.GridNames, side string) (isSpan bool, _ int, _ string, coord maybeInt) {
	isSpan, number, ident := line.IsSpan(), line.Val, line.Ident
	if ident != "" && line.IsCustomIdent() {
		coord.valid = true
		hasBroken := false
		var (
			line pr.GridNames
			tag  = fmt.Sprintf("%s-%s", ident, side)
		)
		for coord.i, line = range lines {
			if utils.IsIn(line, tag) {
				hasBroken = true
				break
			}
		}
		if !hasBroken {
			number = 1
		}
	}
	if number != 0 && !isSpan {
		coord.valid = true
		if ident == "" {
			coord.i = number - 1
		} else {
			step := -1
			if number > 0 {
				step = 1
			}
			L := len(lines) / step
			hasBroken := false
			for coord.i = 0; coord.i < L; coord.i++ {
				line := lines[coord.i*step]
				if utils.IsIn(line, ident) {
					number -= step
					hasBroken = true
					break
				}
				if number == 0 {
					hasBroken = true
					break
				}
			}
			if !hasBroken {
				coord.i += utils.Abs(number)
			}

			if step == -1 {
				coord.i = len(lines) - 1 - coord.i
			}
		}
	}
	if isSpan {
		coord.valid = false
	}
	return isSpan, number, ident, coord
}

type placement [2]int // coord, size

func (pl placement) isNotNone() bool { return pl[1] != 0 }

func (pl placement) unpack() (coord, size int) { return pl[0], pl[1] }

// Input coordinates are 1-indexed, returned coordinates are 0-indexed.
func getPlacement(start, end pr.GridLine, lines []pr.GridNames) placement {
	if start.IsAuto() || start.Tag == pr.Span {
		if end.IsAuto() || end.Tag == pr.Span {
			return placement{}
		}
	}
	var (
		coord            maybeInt
		size             int
		isSpan           bool
		number           int
		spanIdent, ident string
	)
	if start.Tag != pr.Auto {
		isSpan, number, ident, coord = getLine(start, lines, "start")
		if isSpan {
			size = 1
			if number != 0 {
				size = number
			}
			spanIdent = ident
		}
	} else {
		size = 1
		spanIdent = ""
	}
	if end.Tag != pr.Auto {
		var coordEnd maybeInt
		isSpan, number, ident, coordEnd = getLine(end, lines, "end")
		if isSpan {
			size = 1
			if number != 0 {
				size = number
			}
			spanNumber := size
			spanIdent = ident
			if spanIdent != "" {
				hasBroken := false
				for index, line := range lines[coord.i+1:] {
					size = index + 1
					if utils.IsIn(line, spanIdent) {
						spanNumber -= 1
					}
					if spanNumber == 0 {
						hasBroken = true
						break
					}
				}
				if !hasBroken {
					size += spanNumber
				}
			}
		} else if coord.valid {
			size = coordEnd.i - coord.i
		}
		if !coord.valid {
			if spanIdent == "" {
				coord = maybeInt{valid: true, i: coordEnd.i - size}
			} else {
				if number == 0 {
					number = 1
				}
				if coordEnd.i > 0 {
					slice := lines[coordEnd.i-1:]
					hasBroken := false
					for coord.i = range slice {
						line := slice[len(slice)-1-coord.i] // reverse
						if utils.IsIn(line, spanIdent) {
							number -= 1
						}
						if number == 0 {
							coord.i = coordEnd.i - 1 - coord.i
							hasBroken = true
							break
						}
					}
					if !hasBroken {
						coord.i = -number
					}
				} else {
					coord = maybeInt{valid: true, i: -number}
				}
			}
			size = coordEnd.i - coord.i
		}
	} else {
		size = 1
	}

	if size < 0 {
		size = -size
		coord.i -= size
	}
	if size == 0 {
		size = 1
	}
	return placement{coord.i, size}
}

func getSpan(place pr.GridLine) int {
	// TODO: Handle lines.
	span := 1
	if place.IsSpan() && place.Val != 0 {
		span = place.Val
	}
	return span
}

func getColumnPlacement(rowPlacement [2]int, columnStart, columnEnd pr.GridLine,
	columns []pr.GridNames, childrenPositions map[Box]rect, dense bool,
) placement {
	occupiedColumns := map[int]bool{}
	for _, rect := range childrenPositions {
		x, y, width, height := rect.unpack()
		// Test whether cells overlap.
		if intersect(y, height, rowPlacement[0], rowPlacement[1]) {
			for xv := x; xv < x+width; xv++ {
				occupiedColumns[xv] = true
			}
		}
	}
	if dense {
		for x := 0; ; x++ {
			if occupiedColumns[x] {
				continue
			}
			var pl placement
			if columnStart.IsAuto() {
				pl = getPlacement(pr.GridLine{Val: x + 1}, columnEnd, columns)
			} else {
				if columnStart.Tag == pr.Span {
					panic("expected span")
				}
				// If the placement contains two spans, remove the one
				// contributed by the end grid-placement property.
				// https://drafts.csswg.org/css-grid/#grid-placement-errors
				span := getSpan(columnStart)
				pl = getPlacement(columnStart, pr.GridLine{Val: x + 1 + span}, columns)
			}
			hasIntersection := false
			for col := pl[0]; col < pl[0]+pl[1]; col++ {
				if occupiedColumns[col] {
					hasIntersection = true
					break
				}
			}
			if !hasIntersection {
				return pl
			}
		}
	} else {
		y := 0
		for k := range occupiedColumns {
			if k > y {
				y = k
			}
		}
		y += 1
		if columnStart.IsAuto() {
			return getPlacement(pr.GridLine{Val: y + 1}, columnEnd, columns)
		} else {
			if columnStart.Tag == pr.Span {
				panic("expected span")
			}
			// If the placement contains two spans, remove the one contributed
			// by the end grid-placement property.
			// https://drafts.csswg.org/css-grid/#grid-placement-errors
			for endY := y + 1; ; endY++ {
				placement := getPlacement(columnStart, pr.GridLine{Val: endY + 1}, columns)
				if placement[0] >= y {
					return placement
				}
			}
		}
	}
}

const (
	sizeMin byte = iota
	sizeMax
)

// affectedSizes : sizeMin or sizeMax ("min", "max")
// affectedTracksTypes : i, c, m ("intrinsic", "content-based", "max-content")
// sizeContribution : m, c, C ("mininum", "min-content", "max-content")
// direction : x, y
func distributeExtraSpace(context *layoutContext, affectedSizes, affectedTracksTypes, sizeContribution byte, tracksChildren [][]Box,
	sizingFunctions [][2]pr.DimOrS, tracksSizes [][2]pr.MaybeFloat, span int, direction byte, containingBlock *bo.BoxFields,
) {
	// 1. Maintain separately for each affected track a planned increase.
	plannedIncreases := make([]pr.Float, len(tracksSizes))

	// 2. Distribute space.
	affectedTracks := make([]bool, len(sizingFunctions))
	for i, functions := range sizingFunctions {
		function := functions[affectedSizes]
		isAffected := false
		switch affectedTracksTypes {
		case 'i':
			isAffected = function.S == "min-content" || function.S == "max-content" || function.S == "auto"
		case 'c':
			isAffected = function.S == "min-content" || function.S == "max-content"
		case 'm':
			isAffected = function.S == "max-content" || function.S == "auto"
		}
		affectedTracks[i] = isAffected
	}
	for i, children := range tracksChildren {
		if len(children) == 0 {
			continue
		}
		for _, item := range children {
			// 2.1 Find the space distribution.
			// TODO: Differenciate minimum && min-content values.
			// TODO: Find a better way to get height.
			var space pr.Float
			if direction == 'x' {
				if sizeContribution == 'm' || sizeContribution == 'c' {
					space = minContentWidth(context, item, true)
				} else {
					space = maxContentWidth(context, item, true)
				}
			} else {
				item = bo.Deepcopy(item)
				item.Box().PositionX = 0
				item.Box().PositionY = 0
				item, _, _ = blockLevelLayout(context, item.(bo.BlockLevelBoxITF), -pr.Inf, nil,
					containingBlock, true, nil, nil, nil, false, -1)
				space = item.Box().MarginHeight()
			}
			for _, sizes := range tracksSizes[utils.MinInt(i, len(tracksSizes)):utils.MinInt(i+span, len(tracksSizes))] {
				space -= sizes[affectedSizes].V()
			}
			space = pr.Max(0, space)
			// 2.2 Distribute space up to limits.
			var affectedTracksNumbers, unaffectedTracksNumbers []int
			for j := i; j < i+span && j < len(affectedTracks); j++ {
				if affectedTracks[j] {
					affectedTracksNumbers = append(affectedTracksNumbers, j)
				} else {
					unaffectedTracksNumbers = append(unaffectedTracksNumbers, j)
				}
			}
			// tracksNumbers = list(enumerate(affectedTracks[i:i+span], i))
			// affectedTracksNumbers = [                j for j, affected := range tracksNumbers if affected]
			itemIncurredIncreases := make([]pr.Float, len(sizingFunctions))
			distributedSpace := space
			if L := len(affectedTracksNumbers); L != 0 {
				distributedSpace /= pr.Float(L)
			}
			for _, trackNumber := range affectedTracksNumbers {
				// baseSize, growthLimit := tracksSizes[trackNumber]
				itemIncurredIncrease := distributedSpace
				affectedSize := tracksSizes[trackNumber][affectedSizes].V()
				limit := tracksSizes[trackNumber][1].V()
				if affectedSize+itemIncurredIncrease >= limit {
					extra := (itemIncurredIncrease + affectedSize - limit)
					itemIncurredIncrease -= extra
				}
				space -= itemIncurredIncrease
				itemIncurredIncreases[trackNumber] = itemIncurredIncrease
			}
			// 2.3 Distribute space to non-affected tracks.
			if space != 0 && len(affectedTracksNumbers) != 0 {
				// unaffectedTracksNumbers = [                    j for j, affected := range tracksNumbers if ! affected]
				distributedSpace := space
				if L := len(unaffectedTracksNumbers); L != 0 {
					distributedSpace /= pr.Float(L)
				}
				for _, trackNumber := range unaffectedTracksNumbers {
					// baseSize, growthLimit = tracksSizes[trackNumber]
					itemIncurredIncrease := distributedSpace
					affectedSize := tracksSizes[trackNumber][affectedSizes].V()
					limit := tracksSizes[trackNumber][1].V()
					if affectedSize+itemIncurredIncrease >= limit {
						extra := (itemIncurredIncrease + affectedSize - limit)
						itemIncurredIncrease -= extra
					}
					space -= itemIncurredIncrease
					itemIncurredIncreases[trackNumber] = (itemIncurredIncrease)
				}
			}
			// 2.4 Distribute space beyond limits.
			if space != 0 {
				// TODO: Distribute space beyond limits.
			}
			// 2.5. Set the track’s planned increase.
			for k, extra := range itemIncurredIncreases {
				if extra > plannedIncreases[k] {
					plannedIncreases[k] = extra
				}
			}
		}
	}
	// 3. Update the tracks’ affected size.
	for i, increase := range plannedIncreases {
		if affectedSizes == sizeMax && tracksSizes[i][1] == pr.Inf {
			tracksSizes[i][1] = tracksSizes[i][0].V() + increase
		} else {
			tracksSizes[i][affectedSizes] = tracksSizes[i][affectedSizes].V() + increase
		}
	}
}

// direction : 'x' or 'y'
func resolveTracksSizes(context *layoutContext, sizingFunctions [][2]pr.DimOrS, boxSize pr.MaybeFloat, childrenPositions map[Box]rect,
	implicitStart int, direction byte, gap pr.Float,
	containingBlock bo.Box, orthogonalSizes [][2]pr.Float,
) [][2]pr.Float {
	// TODO: Check that auto box size is 0 for percentages.
	percentBoxSize := pr.Float(0)
	if boxSize != pr.AutoF {
		percentBoxSize = boxSize.V()
	}

	// 1.1 Initialize track sizes.
	tracksSizes := make([][2]pr.MaybeFloat, len(sizingFunctions))
	for i, funcs := range sizingFunctions {
		minFunction, maxFunction := funcs[0], funcs[1]
		var baseSize pr.MaybeFloat
		if isLength(minFunction) {
			baseSize = pr.ResolvePercentage(minFunction, percentBoxSize)
		} else if minFunction.S == "min-content" || minFunction.S == "max-content" || minFunction.S == "auto" {
			baseSize = pr.Float(0)
		}
		var growthLimit pr.MaybeFloat
		if isLength(maxFunction) {
			growthLimit = pr.ResolvePercentage(maxFunction, percentBoxSize)
		} else if (maxFunction.S == "min-content" || maxFunction.S == "max-content" || maxFunction.S == "auto") ||
			isFr(maxFunction) {
			growthLimit = pr.Inf
		}
		if baseSize != nil && growthLimit != nil {
			growthLimit = pr.Max(baseSize.V(), growthLimit.V())
		}
		tracksSizes[i] = [2]pr.MaybeFloat{baseSize, growthLimit}
	}

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("resolveTracksSizes(1): %v", tracksSizes))
	}

	// 1.2 Resolve intrinsic track sizes.
	// 1.2.1 Shim baseline-aligned items.
	// TODO: Shim items.
	// 1.2.2 Size tracks to fit non-spanning items.
	tracksChildren := make([][]Box, len(tracksSizes))
	for child, rect := range childrenPositions {
		x, y, width, height := rect.unpack()
		coord, size := y, height
		if direction == 'x' {
			coord, size = x, width
		}
		if size != 1 {
			continue
		}
		tracksChildren[coord-implicitStart] = append(tracksChildren[coord-implicitStart], child)
	}

	for i, children := range tracksChildren {
		minFunction, maxFunction := sizingFunctions[i][0], sizingFunctions[i][1]
		sizes := &tracksSizes[i]
		if len(children) == 0 {
			continue
		}
		if direction == 'y' {
			// TODO: Find a better way to get height.
			height := pr.Float(0)
			for _, child := range children {
				pos := childrenPositions[child]
				x, _, width, _ := pos.unpack()
				widthF := sum0(orthogonalSizes[x : x+width])
				child = bo.Deepcopy(child)
				child.Box().PositionX = 0
				child.Box().PositionY = 0
				parent := bo.BlockT.AnonymousFrom(containingBlock, nil)
				cbW, cbH := containingBlock.Box().ContainingBlock()
				resolvePercentages(parent, bo.MaybePoint{cbW, cbH}, 0)
				parent.Box().PositionX = child.Box().PositionX
				parent.Box().PositionY = child.Box().PositionY
				parent.Box().Width = widthF
				parent.Box().Height = height
				bottomSpace := -pr.Inf
				child, _, _ = blockLevelLayout(context, child.(bo.BlockLevelBoxITF), bottomSpace, nil,
					parent.Box(), true, nil, nil, nil, false, -1)
				height = pr.Max(height, child.Box().MarginHeight())
			}
			if minFunction.S == "min-content" || minFunction.S == "maxContent" || minFunction.S == "auto" {
				sizes[0] = height
			}
			if maxFunction.S == "min-content" || maxFunction.S == "maxContent" {
				sizes[1] = height
			}
			if sizes[0] != nil && sizes[1] != nil {
				sizes[1] = pr.Max(sizes[0].V(), sizes[1].V())
			}
			continue
		}
		if minFunction.S == "min-content" {
			ma := pr.Float(0)
			for _, child := range children {
				if v := minContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[0] = ma
		} else if minFunction.S == "max-content" {
			ma := pr.Float(0)
			for _, child := range children {
				if v := maxContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[0] = ma
		} else if minFunction.S == "auto" {
			// TODO: Handle min-/max-content constrained parents.
			// TODO: Use real "minimum contributions".
			ma := pr.Float(0)
			for _, child := range children {
				if v := minContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[0] = ma
		}
		if maxFunction.S == "min-content" {
			ma := -pr.Inf
			for _, child := range children {
				if v := minContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[1] = ma
		} else if maxFunction.S == "auto" || maxFunction.S == "max-content" {
			ma := -pr.Inf
			for _, child := range children {
				if v := maxContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[1] = ma
			if sizes[0] != nil && sizes[1] != nil {
				sizes[1] = pr.Max(sizes[0].V(), sizes[1].V())
			}
		}
	}

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("resolveTracksSizes(2): %v", tracksSizes))
	}

	// 1.2.3 Increase sizes to accommodate items spanning content-sized tracks.
	var spans []int
	for _, rect := range childrenPositions {
		v := rect[2] // width
		if direction == 'y' {
			v = rect[3] // height
		}
		if v >= 2 {
			spans = append(spans, v)
		}
	}
	sort.Ints(spans)

	for _, span := range spans {
		tracksChildren := make([][]Box, len(sizingFunctions))
		i := -1
		for child, rect := range childrenPositions {
			i++
			x, y, width, height := rect.unpack()
			coord, size := x, width
			if direction == 'y' {
				coord, size = y, height
			}
			if size != span {
				continue
			}
			hasFr := false
			for _, functions := range sizingFunctions[utils.MinInt(i, len(sizingFunctions)):utils.MinInt(len(sizingFunctions), i+span+1)] {
				if isFr(functions[1]) {
					hasFr = true
					break
				}
			}
			if !hasFr {
				tracksChildren[coord-implicitStart] = append(tracksChildren[coord-implicitStart], child)
			}
		}
		// 1.2.3.1 For intrinsic minimums.
		// TODO: Respect min-/max-content constraint.
		distributeExtraSpace(context, sizeMin, 'i', 'm', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock.Box())
		// 1.2.3.2 For content-based minimums.
		distributeExtraSpace(context, sizeMin, 'c', 'c', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock.Box())
		// 1.2.3.3 For max-content minimums.
		// TODO: Respect max-content constraint.
		distributeExtraSpace(context, sizeMin, 'm', 'C', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock.Box())
		// 1.2.3.4 Increase growth limit.
		for j, sizes := range tracksSizes {
			if sizes[0] != nil && sizes[1] != nil {
				tracksSizes[j][1] = pr.Max(sizes[0].V(), sizes[1].V())
			}
		}
		i = -1
		for child, rect := range childrenPositions {
			i++
			x, y, width, height := rect.unpack()
			coord, size := x, width
			if direction == 'y' {
				coord, size = y, height
			}
			if size != span {
				continue
			}

			hasFr := false
			for _, functions := range sizingFunctions[utils.MinInt(i, len(sizingFunctions)):utils.MinInt(len(sizingFunctions), i+span+1)] {
				if isFr(functions[1]) {
					hasFr = true
					break
				}
			}
			if !hasFr {
				tracksChildren[coord-implicitStart] = append(tracksChildren[coord-implicitStart], child)
			}
		}
		// 1.2.3.5 For intrinsic maximums.
		distributeExtraSpace(context, sizeMax, 'i', 'c', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock.Box())
		// 1.2.3.6 For max-content maximums.
		distributeExtraSpace(context, sizeMax, 'm', 'C', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock.Box())
	}
	// 1.2.4 Increase sizes to accommodate items spanning flexible tracks.
	// TODO: Support spans for flexible tracks.
	// 1.2.5 Fix infinite growth limits.
	for i, sizes := range tracksSizes {
		if sizes[1] == pr.Inf {
			tracksSizes[i][1] = sizes[0]
		}
	}
	// 1.3 Maximize tracks.
	var freeSpace pr.MaybeFloat
	if boxSize != pr.AutoF {
		sum := pr.Float(0)
		for _, size := range tracksSizes {
			sum += size[0].V()
		}

		freeSpaceF := boxSize.V() - sum - pr.Float(len(tracksSizes)-1)*gap
		if freeSpaceF > 0 {
			distributedFreeSpace := freeSpaceF / pr.Float(len(tracksSizes))
			for i := range tracksSizes {
				sizes := &tracksSizes[i]
				baseSize, growthLimit := sizes[0].V(), sizes[1].V()
				if baseSize+distributedFreeSpace > growthLimit {
					sizes[0] = growthLimit
					freeSpaceF -= growthLimit - baseSize
				} else {
					sizes[0] = baseSize + distributedFreeSpace
					freeSpaceF -= distributedFreeSpace
				}
			}
		}
		freeSpace = freeSpaceF
	}
	// TODO: Respect max-width/-height.
	// 1.4 Expand flexible tracks.
	var flexFraction pr.Float
	if freeSpace != nil && freeSpace.V() <= 0 {
		// TODO: Respect min-content constraint.
		flexFraction = 0
	} else if freeSpace != nil {
		stop := false
		inflexibleTracks := make([]bool, len(tracksSizes))
		var hypotheticalFrSize pr.Float
		for !stop {
			leftoverSpace := freeSpace.V()
			flexFactorSum := pr.Float(0)
			for i, sizes := range tracksSizes {
				maxFunction := sizingFunctions[i][1]
				if isFr(maxFunction) {
					leftoverSpace += sizes[0].V()
					if !inflexibleTracks[i] {
						flexFactorSum += maxFunction.Value
					}
				}
			}
			flexFactorSum = pr.Max(1, flexFactorSum)
			hypotheticalFrSize = leftoverSpace / flexFactorSum
			stop = true
			// iterable = enumerate(zip(tracksSizes, sizingFunctions))
			for i, sizes := range tracksSizes {
				maxFunction := sizingFunctions[i][1]
				if !inflexibleTracks[i] && isFr(maxFunction) {
					if hypotheticalFrSize*maxFunction.Value < sizes[0].V() {
						inflexibleTracks[i] = true
						stop = false
					}
				}
			}
		}
		flexFraction = hypotheticalFrSize
	} else {
		flexFraction = 0
		for i, sizes := range tracksSizes {
			maxFunction := sizingFunctions[i][1]
			if isFr(maxFunction) {
				if maxFunction.Value > 1 {
					flexFraction = pr.Max(flexFraction, maxFunction.Value*sizes[0].V())
				} else {
					flexFraction = pr.Max(flexFraction, sizes[0].V())
				}
			}
		}
		// TODO: Respect grid items max-content contribution.
		// TODO: Respect min-* constraint.
	}
	for i, sizes := range tracksSizes {
		maxFunction := sizingFunctions[i][1]
		if isFr(maxFunction) {
			if flexFraction*maxFunction.Value > sizes[0].V() {
				if freeSpace != nil {
					freeSpace = freeSpace.V() - flexFraction*maxFunction.Value
				}
				tracksSizes[i][0] = flexFraction * maxFunction.Value
			}
		}
	}
	// 1.5 Expand stretched auto tracks.
	justifyContent := containingBlock.Box().Style.GetJustifyContent()
	alignContent := containingBlock.Box().Style.GetAlignContent()
	xStretch := direction == 'x' && justifyContent.Intersects(kw.Normal, kw.Stretch)
	yStretch := direction == 'y' && alignContent.Intersects(kw.Normal, kw.Stretch)
	if (xStretch || yStretch) && freeSpace != nil && freeSpace.V() > 0 {
		var autoTracksSizes []*[2]pr.MaybeFloat
		for i := range tracksSizes {
			minFunction := sizingFunctions[i][0]
			if minFunction.S == "auto" {
				autoTracksSizes = append(autoTracksSizes, &tracksSizes[i])
			}
		}
		if len(autoTracksSizes) != 0 {
			distributedFreeSpace := freeSpace.V() / pr.Float(len(autoTracksSizes))
			for _, ptr := range autoTracksSizes {
				ptr[0] = ptr[0].V() + distributedFreeSpace
			}
		}
	}

	out := make([][2]pr.Float, len(tracksSizes))
	for i, m := range tracksSizes {
		out[i] = [2]pr.Float{m[0].V(), m[1].V()}
	}
	return out
}

// return the equivalent of Python l[::2]
func extractNames(rows []pr.GridSpec) []pr.GridNames {
	var names []pr.GridNames
	for i, row := range rows {
		if i%2 != 0 {
			continue
		}
		names = append(names, row.(pr.GridNames))
	}
	return names
}

// return the equivalent of Python l[1::2]
func extractDims(rows []pr.GridSpec) [][2]pr.DimOrS {
	var dims [][2]pr.DimOrS
	for i, row := range rows {
		if i%2 != 1 {
			continue
		}
		dims = append(dims, row.(pr.GridDims).SizingFunctions())
	}
	return dims
}

func sum0(l [][2]pr.Float) pr.Float {
	su := pr.Float(0)
	for _, size := range l {
		su += size[0]
	}
	return su
}

type rect [4]int

func (ir rect) unpack() (x, y, width, height int) {
	return ir[0], ir[1], ir[2], ir[3]
}

func findBoxIndex(l []Box, v Box) int {
	for i, b := range l {
		if b == v {
			return i
		}
	}
	return -1
}

func gridLayout(context *layoutContext, box_ Box, bottomSpace pr.Float, skipStack tree.ResumeStack, containingBlock containingBlock,
	pageIsEmpty bool, absoluteBoxes, fixedBoxes *[]*AbsolutePlaceholder,
) (Box, blockLayout) {
	context.createBlockFormattingContext()

	// Define explicit grid
	box, style := box_.Box(), box_.Box().Style
	gridAreas := style.GetGridTemplateAreas()
	flow := style.GetGridAutoFlow()
	autoRows := style.GetGridAutoRows().Cycle()
	autoColumns := style.GetGridAutoColumns().Cycle()
	autoRowsBack := style.GetGridAutoRows().Reverse().Cycle()
	autoColumnsBack := style.GetGridAutoColumns().Reverse().Cycle()
	var rowGap, columnGap pr.Float
	if v := style.GetColumnGap(); v.S != "normal" {
		columnGap = v.Value
	}
	if v := style.GetRowGap(); v.S != "normal" {
		rowGap = v.Value
	}

	// TODO: Support "column" value in grid-auto-flow.
	if utils.IsIn(flow, "column") {
		logger.WarningLogger.Println(`"column" is not supported in grid-auto-flow`)
	}

	if gridAreas.IsNone() {
		gridAreas = pr.GridTemplateAreas{{""}}
	}

	rows := getTemplateTracks(style.GetGridTemplateRows())
	columns := getTemplateTracks(style.GetGridTemplateColumns())

	// Adjust rows number
	gridAreasColumns := 0
	if len(gridAreas) != 0 {
		gridAreasColumns = len(gridAreas[0])
	}
	rowsDiff := (len(rows)-1)/2 - len(gridAreas)
	if rowsDiff > 0 {
		for c := 0; c < rowsDiff; c++ {
			gridAreas = append(gridAreas, make([]string, gridAreasColumns))
		}
	} else if rowsDiff < 0 {
		for c := 0; c < -rowsDiff; c++ {
			rows = append(rows, autoRows.Next(), pr.GridNames{})
		}
	}

	// Adjust columns number
	columnsDiff := (len(columns)-1)/2 - gridAreasColumns
	if columnsDiff > 0 {
		for i := range gridAreas {
			gridAreas[i] = append(gridAreas[i], make([]string, columnsDiff)...)
		}
	} else if columnsDiff < 0 {
		for c := 0; c < -columnsDiff; c++ {
			columns = append(columns, autoColumns.Next(), pr.GridNames{})
		}
	}

	// Add implicit line names
	for y, row := range gridAreas {
		for x, areaName := range row {
			if areaName == "" {
				continue
			}
			startName := areaName + "-start"
			var names []string
			for _, row := range extractNames(rows) {
				names = append(names, row...)
			}
			if !utils.IsIn(names, startName) {
				rows[2*y] = append(rows[2*y].(pr.GridNames), startName)
			}
			names = names[:0]
			for _, column := range extractNames(columns) {
				names = append(names, column...)
			}
			if !utils.IsIn(names, startName) {
				columns[2*x] = append(columns[2*x].(pr.GridNames), startName)
			}
		}
	}
	for y := range gridAreas {
		row := gridAreas[len(gridAreas)-1-y] // reverse
		for x := range row {
			areaName := row[len(row)-1-x] // reverse
			if areaName == "" {
				continue
			}
			endName := areaName + "-end"

			var names []string
			for _, row := range extractNames(rows) {
				names = append(names, row...)
			}
			if !utils.IsIn(names, endName) {
				rows[len(rows)-2*y-1] = append(rows[len(rows)-2*y-1].(pr.GridNames), endName)
			}
			names = names[:0]
			for _, column := range extractNames(columns) {
				names = append(names, column...)
			}
			if !utils.IsIn(names, endName) {
				columns[len(columns)-2*x-1] = append(columns[len(columns)-2*x-1].(pr.GridNames), endName)
			}
		}
	}

	// 1. Run the grid placement algorithm.

	// 1.1 Position anything that’s not auto-positioned.
	childrenPositions := map[Box]rect{}
	for _, child := range box.Children {
		columnStart := child.Box().Style.GetGridColumnStart()
		columnEnd := child.Box().Style.GetGridColumnEnd()
		rowStart := child.Box().Style.GetGridRowStart()
		rowEnd := child.Box().Style.GetGridRowEnd()

		columnPlacement := getPlacement(columnStart, columnEnd, extractNames(columns))
		rowPlacement := getPlacement(rowStart, rowEnd, extractNames(rows))
		if columnPlacement.isNotNone() && rowPlacement.isNotNone() {
			x, width := columnPlacement.unpack()
			y, height := rowPlacement.unpack()
			childrenPositions[child] = rect{x, y, width, height}
		}
	}

	// 1.2 Process the items locked to a given row.
	children := make([]Box, len(box.Children))
	copy(children, box.Children)
	sort.Slice(children, func(i, j int) bool { return children[i].Box().Style.GetOrder() < children[j].Box().Style.GetOrder() })
	for _, child := range children {
		if _, has := childrenPositions[child]; has {
			continue
		}
		rowStart := child.Box().Style.GetGridRowStart()
		rowEnd := child.Box().Style.GetGridRowEnd()
		rowPlacement := getPlacement(rowStart, rowEnd, extractNames(rows))
		if !rowPlacement.isNotNone() {
			continue
		}
		y, height := rowPlacement[0], rowPlacement[1]
		columnStart := child.Box().Style.GetGridColumnStart()
		columnEnd := child.Box().Style.GetGridColumnEnd()
		x, width := getColumnPlacement(rowPlacement, columnStart, columnEnd, extractNames(columns),
			childrenPositions, utils.IsIn(flow, "dense")).unpack()
		childrenPositions[child] = [4]int{x, y, width, height}
	}

	// 1.3 Determine the columns in range the implicit grid.
	// 1.3.1 Start with the columns from the explicit grid.
	implicitX1 := 0
	implicitX2 := 0
	if len(gridAreas) != 0 {
		implicitX2 = len(gridAreas[0])
	}
	// 1.3.2 Add columns to the beginning and end of the implicit grid.
	var remainingGridItems []Box
	for _, child := range children {
		var x, width int
		if r, has := childrenPositions[child]; has {
			x, _, width, _ = r.unpack()
		} else {
			columnStart := child.Box().Style.GetGridColumnStart()
			columnEnd := child.Box().Style.GetGridColumnEnd()
			columnPlacement := getPlacement(columnStart, columnEnd, extractNames(columns))
			remainingGridItems = append(remainingGridItems, child)
			if columnPlacement.isNotNone() {
				x, width = columnPlacement.unpack()
			} else {
				continue
			}
		}
		implicitX1 = utils.MinInt(x, implicitX1)
		implicitX2 = utils.MaxInt(x+width, implicitX2)
	}
	// 1.3.3 Add columns to accommodate max column span.
	for _, child := range remainingGridItems {
		columnStart := child.Box().Style.GetGridColumnStart()
		columnEnd := child.Box().Style.GetGridColumnEnd()
		span := 1
		if columnStart.IsSpan() {
			span = columnStart.Val
		} else if columnEnd.IsSpan() {
			span = columnEnd.Val
		}
		if span == 0 {
			span = 1
		}
		implicitX2 = utils.MaxInt(implicitX1+span, implicitX2)
	}

	// 1.4 Position the remaining grid items.
	implicitY1 := 0
	implicitY2 := len(gridAreas)
	for _, position := range childrenPositions {
		_, y, _, height := position.unpack()
		implicitY1 = utils.MinInt(y, implicitY1)
		implicitY2 = utils.MaxInt(y+height, implicitY2)
	}
	cursorX, cursorY := implicitX1, implicitY1
	if utils.IsIn(flow, "dense") {
		for _, child := range remainingGridItems {
			columnStart := child.Box().Style.GetGridColumnStart()
			columnEnd := child.Box().Style.GetGridColumnEnd()
			columnPlacement := getPlacement(columnStart, columnEnd, extractNames(columns))
			if columnPlacement.isNotNone() {
				// 1. Set the row position of the cursor.
				cursorY = implicitY1
				x, width := columnPlacement[0], columnPlacement[1]
				cursorX = x
				// 2. Increment the cursor’s row position.
				rowStart := child.Box().Style.GetGridRowStart()
				rowEnd := child.Box().Style.GetGridRowEnd()
				var y, height int
				for y = cursorY; ; y++ {
					if rowStart.IsAuto() {
						y, height = getPlacement(pr.GridLine{Val: y + 1}, rowEnd, extractNames(rows)).unpack()
					} else {
						// assert rowStart[0] == "span"
						// assert rowStart.IsAuto() || rowStart[0] == "span"
						span := getSpan(rowStart)
						y, height = getPlacement(rowStart, pr.GridLine{Val: y + 1 + span}, extractNames(rows)).unpack()
					}
					if y < cursorY {
						continue
					}
					hasBroken := false
					for row := y; row < y+height; row++ {
						intersect := intersectWithChildren(x, y, width, height, childrenPositions)
						if intersect {
							// Child intersects with a positioned child on
							// current row.
							hasBroken = true
							break
						}
					}
					if !hasBroken {
						// Child doesn’t intersect with any positioned child on
						// any row.
						break
					}
				}
				yDiff := y + height - implicitY2
				if yDiff > 0 {
					for c := 0; c < yDiff; c++ {
						rows = append(rows, autoRows.Next(), pr.GridNames{})
					}
					implicitY2 = y + height
				}
				// 3. Set the item’s row-start line.
				childrenPositions[child] = rect{x, y, width, height}
			} else {
				// 1. Set the cursor’s row && column positions.
				cursorX, cursorY = implicitX1, implicitY1
				for {
					// 2. Increment the column position of the cursor.
					y := cursorY
					rowStart := child.Box().Style.GetGridRowStart()
					rowEnd := child.Box().Style.GetGridRowEnd()
					columnStart = child.Box().Style.GetGridColumnStart()
					columnEnd = child.Box().Style.GetGridColumnEnd()
					hasBroken := false
					for x := cursorX; x < implicitX2; x++ {
						var width, height int
						if rowStart.IsAuto() {
							y, height = getPlacement(pr.GridLine{Val: y + 1}, rowEnd, extractNames(rows)).unpack()
						} else {
							span := getSpan(rowStart)
							y, height = getPlacement(rowStart, pr.GridLine{Val: y + 1 + span}, extractNames(rows)).unpack()
						}
						if columnStart.IsAuto() {
							x, width = getPlacement(pr.GridLine{Val: x + 1}, columnEnd, extractNames(columns)).unpack()
						} else {
							span := getSpan(columnStart)
							x, width = getPlacement(columnStart, pr.GridLine{Val: x + 1 + span}, extractNames(columns)).unpack()
						}
						intersect := intersectWithChildren(x, y, width, height, childrenPositions)
						if intersect {
							// Child intersects with a positioned child.
							continue
						} else {
							// Free place found.
							// 3. Set the item’s row-/column-start lines.
							childrenPositions[child] = [4]int{x, y, width, height}
							yDiff := cursorY + height - 1 - implicitY2
							if yDiff > 0 {
								for c := 0; c < yDiff; c++ {
									rows = append(rows, autoRows.Next(), pr.GridNames{})
								}
								implicitY2 = cursorY + height - 1
							}
							hasBroken = true
							break
						}
					}
					if !hasBroken {
						// No room found.
						// 2. Return to the previous step.
						cursorY += 1
						yDiff := cursorY + 1 - implicitY2
						if yDiff > 0 {
							for c := 0; c < yDiff; c++ {
								rows = append(rows, autoRows.Next(), pr.GridNames{})
							}
							implicitY2 = cursorY
						}
						cursorX = implicitX1
						continue
					}
					break
				}
			}
		}
	} else {
		for _, child := range remainingGridItems {
			columnStart := child.Box().Style.GetGridColumnStart()
			columnEnd := child.Box().Style.GetGridColumnEnd()
			columnPlacement := getPlacement(columnStart, columnEnd, extractNames(columns))
			if columnPlacement.isNotNone() {
				// 1. Set the column position of the cursor.
				x, width := columnPlacement.unpack()
				if x < cursorX {
					cursorY += 1
				}
				cursorX = x
				// 2. Increment the cursor’s row position.
				rowStart := child.Box().Style.GetGridRowStart()
				rowEnd := child.Box().Style.GetGridRowEnd()
				var y, height int
				for ; ; cursorY++ {
					if rowStart.IsAuto() {
						y, height = getPlacement(pr.GridLine{Val: cursorY + 1}, rowEnd, extractNames(rows)).unpack()
					} else {
						// assert rowStart[0] == "span"
						// assert rowStart.IsAuto() || rowStart[0] == "span"
						span := getSpan(rowStart)
						y, height = getPlacement(rowStart, pr.GridLine{Val: cursorY + 1 + span}, extractNames(rows)).unpack()
					}
					if y < cursorY {
						continue
					}
					hasBroken := false
					for row := y; row < y+height; row++ {
						intersect := intersectWithChildren(x, y, width, height, childrenPositions)
						if intersect {
							// Child intersects with a positioned child on
							// current row.
							hasBroken = true
							break
						}
					}
					if !hasBroken {
						// Child doesn’t intersect with any positioned child on
						// any row.
						break
					}
				}
				yDiff := y + height - implicitY2
				if yDiff > 0 {
					for c := 0; c < yDiff; c++ {
						rows = append(rows, autoRows.Next(), pr.GridNames{})
					}
					implicitY2 = y + height
				} // 3. Set the item’s row-start line.
				childrenPositions[child] = [4]int{x, y, width, height}
			} else {
				for {
					// 1. Increment the column position of the cursor.
					y := cursorY
					rowStart := child.Box().Style.GetGridRowStart()
					rowEnd := child.Box().Style.GetGridRowEnd()
					columnStart = child.Box().Style.GetGridColumnStart()
					columnEnd = child.Box().Style.GetGridColumnEnd()
					hasBroken := false
					for x := cursorX; x < implicitX2; x++ {
						var width, height int
						if rowStart.IsAuto() {
							y, height = getPlacement(pr.GridLine{Val: y + 1}, rowEnd, extractNames(rows)).unpack()
						} else {
							span := getSpan(rowStart)
							y, height = getPlacement(rowStart, pr.GridLine{Val: y + 1 + span},
								extractNames(rows)).unpack()
						}
						if columnStart.IsAuto() {
							x, width = getPlacement(pr.GridLine{Val: x + 1}, columnEnd, extractNames(columns)).unpack()
						} else {
							span := getSpan(columnStart)
							x, width = getPlacement(columnStart, pr.GridLine{Val: x + 1 + span},
								extractNames(columns)).unpack()
						}
						intersect := intersectWithChildren(x, y, width, height, childrenPositions)
						if intersect {
							// Child intersects with a positioned child.
							continue
						} else {
							// Free place found.
							// 2. Set the item’s row-/column-start lines.
							childrenPositions[child] = [4]int{x, y, width, height}
							hasBroken = true
							break
						}
					}
					if !hasBroken {
						// No room found.
						// 2. Return to the previous step.
						cursorY += 1
						yDiff := cursorY + 1 - implicitY2
						if yDiff > 0 {
							for c := 0; c < yDiff; c++ {
								rows = append(rows, autoRows.Next(), pr.GridNames{})
							}
							implicitY2 = cursorY
						}
						cursorX = implicitX1
						continue
					}
					break
				}
			}
		}
	}

	if traceMode {
		traceLogger.Dump(fmt.Sprintf("gridLayout: childrenPositions=%v", childrenPositions))
	}

	for c := 0; c < -implicitX1; c++ {
		columns = append([]pr.GridSpec{pr.GridNames{}, autoColumnsBack.Next()}, columns...)
	}
	start := 0
	if len(gridAreas) != 0 {
		start = len(gridAreas[0])
	}
	for c := start; c < implicitX2; c++ {
		columns = append(columns, autoColumns.Next(), pr.GridNames{})
	}
	for c := 0; c < -implicitY1; c++ {
		rows = append([]pr.GridSpec{pr.GridNames{}, autoRowsBack.Next()}, rows...)
	}
	for c := len(gridAreas); c < implicitY2; c++ {
		rows = append(rows, autoRows.Next(), pr.GridNames{})
	}

	// 2. Find the size of the grid container.

	if bo.GridT.IsInstance(box_) {
		blockLevelWidth(box_, context, containingBlock)
	} else {
		// assert isinstance(box, boxes.InlineGridBox)
		// from .inline import inlineBlockWidth
		inlineBlockWidth(box_, context, containingBlock)
	}
	if box.Width == pr.AutoF {
		// TODO: Calculate max-width.
		box.Width, _ = containingBlock.ContainingBlock()
	}

	// 3. Run the grid sizing algorithm.

	// 3.0 List min/max sizing functions.
	rowSizingFunctions := extractDims(rows)
	columnSizingFunctions := extractDims(columns)

	// 3.1 Resolve the sizes of the grid columns.
	columnsSizes := resolveTracksSizes(context, columnSizingFunctions, box.Width, childrenPositions, implicitX1,
		'x', columnGap, box_, nil)

	// 3.2 Resolve the sizes of the grid rows.
	rowsSizes := resolveTracksSizes(context, rowSizingFunctions, box.Height, childrenPositions, implicitY1,
		'y', rowGap, box_, columnsSizes)

	// 3.3 Re-resolve the sizes of the grid columns with min-/max-content.
	// TODO: Re-resolve.

	// 3.4 Re-re-resolve the sizes of the grid columns with min-/max-content.
	// TODO: Re-re-resolve.

	// 3.5 Align the tracks within the grid container.
	// TODO: Support safe/unsafe.
	justifyContent := box.Style.GetJustifyContent()
	x := box.ContentBoxX()

	freeWidth := pr.Max(0, box.Width.V()-sum0(columnsSizes))
	columnsNumber := pr.Float(len(columnsSizes))
	columnsPositions := make([]pr.Float, len(columnsSizes))
	if justifyContent.Intersects(kw.Center) {
		x += freeWidth / 2
		for i, size := range columnsSizes {
			columnsPositions[i] = x
			x += size[0] + columnGap
		}
	} else if justifyContent.Intersects(kw.Right, kw.End, kw.FlexEnd) {
		x += freeWidth
		for i, size := range columnsSizes {
			columnsPositions[i] = x
			x += size[0] + columnGap
		}
	} else if justifyContent.Intersects(kw.SpaceAround) {
		x += freeWidth / 2 / columnsNumber
		for i, size := range columnsSizes {
			columnsPositions[i] = x
			x += size[0] + freeWidth/columnsNumber + columnGap
		}
	} else if justifyContent.Intersects(kw.SpaceBetween) {
		for i, size := range columnsSizes {
			columnsPositions[i] = x
			if columnsNumber >= 2 {
				x += size[0] + freeWidth/(columnsNumber-1) + columnGap
			}
		}
	} else if justifyContent.Intersects(kw.SpaceEvenly) {
		x += freeWidth / (columnsNumber + 1)
		for i, size := range columnsSizes {
			columnsPositions[i] = x
			x += size[0] + freeWidth/(columnsNumber+1) + columnGap
		}
	} else {
		for i, size := range columnsSizes {
			columnsPositions[i] = x
			x += size[0] + columnGap
		}
	}

	alignContent := box.Style.GetAlignContent()
	y := box.ContentBoxY()
	freeHeight := pr.Float(0)
	if h := box.Height; h != pr.AutoF {
		freeHeight = h.V() - sum0(rowsSizes) - pr.Float(len(rowsSizes)-1)*rowGap
		freeHeight = pr.Max(0, freeHeight)
	}
	rowsNumber := pr.Float(len(rowsSizes))
	rowsPositions := make([]pr.Float, len(rowsSizes))
	if alignContent.Intersects(kw.Center) {
		y += freeHeight / 2
		for i, size := range rowsSizes {
			rowsPositions[i] = y
			y += size[0] + rowGap
		}
	} else if alignContent.Intersects(kw.Right, kw.End, kw.FlexEnd) {
		y += freeHeight
		for i, size := range rowsSizes {
			rowsPositions[i] = y
			y += size[0] + rowGap
		}
	} else if alignContent.Intersects(kw.SpaceAround) {
		y += freeHeight / 2 / rowsNumber
		for i, size := range rowsSizes {
			rowsPositions[i] = y
			y += size[0] + freeHeight/rowsNumber + rowGap
		}
	} else if alignContent.Intersects(kw.SpaceBetween) {
		for i, size := range rowsSizes {
			rowsPositions[i] = y
			if rowsNumber >= 2 {
				y += size[0] + freeHeight/(rowsNumber-1) + rowGap
			}
		}
	} else if alignContent.Intersects(kw.SpaceEvenly) {
		y += freeHeight / (rowsNumber + 1)
		for i, size := range rowsSizes {
			rowsPositions[i] = y
			y += size[0] + freeHeight/(rowsNumber+1) + rowGap
		}
	} else {
		if alignContent.Intersects(kw.Baseline) {
			// TODO: Support baseline value.
			logger.WarningLogger.Println("Baseline alignment is not supported for grid layout")
		}
		for i, size := range rowsSizes {
			rowsPositions[i] = y
			y += size[0] + rowGap
		}
	}

	// 4. Lay out the grid items into their respective containing blocks.
	// Find resumeAt row.
	var (
		thisPageChildren    []Box
		resumeRow           = -1
		skipRow, skipHeight = 0, pr.Float(0)
		resumeAt            tree.ResumeStack
	)
	if skipStack != nil {
		skipRow, _ = skipStack.Unpack()
		skipHeight = (sum0(rowsSizes[:skipRow]) + pr.Float(len(rowsSizes[:skipRow])-1)*rowGap)
	}
	hasBroken := false
	for i := skipRow; i < len(rowsPositions); i++ {
		rowY := rowsPositions[i]
		// TODO: Check that page is not empty.
		if context.overflowsPage(bottomSpace, rowY-skipHeight) {
			if i == 0 {
				return nil, blockLayout{nil, nil, tree.PageBreak{Break: "any"}, false}
			}
			resumeRow = i - 1
			resumeAt = tree.ResumeStack{i - 1: nil}
			for _, child := range children {
				_, y, _, _ := childrenPositions[child].unpack()
				if skipRow <= y && y <= i-2 {
					thisPageChildren = append(thisPageChildren, child)
				}
			}
			hasBroken = true
			break
		}
	}
	if !hasBroken {
		for _, child := range children {
			_, y, _, _ := childrenPositions[child].unpack()
			if skipRow <= y {
				thisPageChildren = append(thisPageChildren, child)
			}
		}
	}
	if box.Height == pr.AutoF {
		slice := resumeRow
		if resumeRow == -1 {
			slice = len(rowsSizes)
		}
		box.Height = sum0(rowsSizes[skipRow:slice]) + pr.Float(len(rowsSizes[skipRow:slice])-1)*rowGap
	}
	// Lay out grid items.
	justifyItems := box.Style.GetJustifyItems()
	alignItems := box.Style.GetAlignItems()
	var (
		newChildren []Box
		baseline    pr.MaybeFloat
	)
	for _, child := range thisPageChildren {
		x, y, width, height := childrenPositions[child].unpack()
		index := findBoxIndex(box.Children, child)
		var childSkipStack tree.ResumeStack
		if v, has := skipStack[y][index]; has {
			childSkipStack = v
		}
		child = bo.Deepcopy(child)
		childB := child.Box()
		childB.PositionX = columnsPositions[x]
		childB.PositionY = rowsPositions[y] - skipHeight
		cbW, cbH := box.ContainingBlock()
		resolvePercentages(child, bo.MaybePoint{cbW, cbH}, 0)
		widthF := (sum0(columnsSizes[x:x+width]) + pr.Float(width-1)*columnGap)
		heightF := (sum0(rowsSizes[y:utils.MinInt(y+height, len(rowsSizes))]) + pr.Float(height-1)*rowGap)
		childWidth := widthF - (childB.MarginLeft.V() + childB.BorderLeftWidth + childB.PaddingLeft.V() +
			childB.MarginRight.V() + childB.BorderRightWidth + childB.PaddingRight.V())
		childHeight := heightF - (childB.MarginTop.V() + childB.BorderTopWidth + childB.PaddingTop.V() +
			childB.MarginBottom.V() + childB.BorderBottomWidth + childB.PaddingBottom.V())

		justifySelf := childB.Style.GetJustifySelf()
		if justifySelf.Intersects(kw.Auto) {
			justifySelf = justifyItems
		}
		if justifySelf.Intersects(kw.Normal, kw.Stretch) {
			if childB.Style.GetWidth().S == "auto" {
				childB.Style.SetWidth(pr.FToPx(childWidth))
			}
		}
		alignSelf := childB.Style.GetAlignSelf()
		if alignSelf.Intersects(kw.Auto) {
			alignSelf = alignItems
		}
		if alignSelf.Intersects(kw.Normal, kw.Stretch) {
			if childB.Style.GetHeight().S == "auto" {
				childB.Style.SetHeight(pr.FToPx(childHeight))
			}
		}

		// TODO: Find a better solution for the layout.
		parent := bo.BlockT.AnonymousFrom(box_, nil)
		cbW, cbH = containingBlock.ContainingBlock()
		resolvePercentages(parent, bo.MaybePoint{cbW, cbH}, 0)
		parent.Box().PositionX = childB.PositionX
		parent.Box().PositionY = childB.PositionY
		parent.Box().Width = widthF
		parent.Box().Height = heightF
		newChild, _, _ := blockLevelLayout(context, child.(bo.BlockLevelBoxITF), bottomSpace, childSkipStack, parent.Box(),
			pageIsEmpty, absoluteBoxes, fixedBoxes, nil, false, -1)
		if newChild != nil {
			pageIsEmpty = false
			// TODO: Support fragmentation in grid items.
		} else {
			// TODO: Support fragmentation in grid rows.
			continue
		}

		// TODO: Apply auto margins.
		if justifySelf.Intersects(kw.Normal, kw.Stretch) {
			newChild.Box().Width = pr.Max(childWidth, newChild.Box().Width.V())
		} else {
			newChild.Box().Width = maxContentWidth(context, newChild, true)
			diff := childWidth - newChild.Box().Width.V()
			if justifySelf.Intersects(kw.Center) {
				newChild.Translate(newChild, diff/2, 0, false)
			} else if justifySelf.Intersects(kw.Right, kw.End, kw.FlexEnd, kw.SelfEnd) {
				newChild.Translate(newChild, diff, 0, false)
			}
		}

		// TODO: Apply auto margins.
		if alignSelf.Intersects(kw.Normal, kw.Stretch) {
			newChild.Box().Height = pr.Max(childHeight, newChild.Box().Height.V())
		} else {
			diff := childHeight - newChild.Box().Height.V()
			if alignSelf.Intersects(kw.Center) {
				newChild.Translate(newChild, 0, diff/2, false)
			} else if alignSelf.Intersects(kw.End, kw.FlexEnd, kw.SelfEnd) {
				newChild.Translate(newChild, 0, diff, false)
			}
		}

		// TODO: Take care of page fragmentation.
		newChildren = append(newChildren, newChild)
		if baseline == nil && y == implicitY1 {
			baseline = findInFlowBaseline(newChild, false)
		}
	}

	box_ = bo.CopyWithChildren(box_, newChildren)
	if bo.InlineGridT.IsInstance(box_) {
		// TODO: Synthetize a real baseline value.
		logger.WarningLogger.Println("Inline grids are not supported")
		if !pr.Is(baseline) {
			baseline = pr.Float(0)
		}
		box_.Box().Baseline = baseline
	}

	context.finishBlockFormattingContext(box_)

	if traceMode {
		traceLogger.DumpTree(box_, "after gridLayout")
	}

	return box_, blockLayout{resumeAt, nil, tree.PageBreak{Break: "any"}, false}
}
