package layout

import (
	"fmt"
	"sort"

	pr "github.com/benoitkugler/webrender/css/properties"
	bo "github.com/benoitkugler/webrender/html/boxes"
	"github.com/benoitkugler/webrender/logger"
	"github.com/benoitkugler/webrender/utils"
)

// Layout for grid containers and grid-items.

func isLength(sizing pr.DimOrS) bool { return sizing.Unit != pr.Fr }

func isFr(sizing pr.DimOrS) bool { return sizing.Unit == pr.Fr }

func intersect(position1, size1, position2, size2 int) bool {
	return position1 < position2+size2 && position2 < position1+size1
}

func intersectWithChildren(x, y, width, height int, positions [][4]int) bool {
	for _, rect := range positions {
		fullX, fullY, fullWidth, fullHeight := rect[0], rect[1], rect[2], rect[3]
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
			if track, isRepeat := track.(pr.GridRepeat); isRepeat {
				repeatNumber, repeatTrackList := track.Repeat, track.Names
				if repeatNumber == pr.RepeatAutoFill || repeatNumber == pr.RepeatAutoFit {
					// TODO: Respect auto-fit && auto-fill.
					logger.WarningLogger.Println(`"auto-fit" and "auto-fill" are unsupported in repeat()`)
					repeatNumber = 1
				}
				for _c := 0; _c < repeatNumber; _c++ {
					for j, repeatTrack := range repeatTrackList {
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

func getLine(line pr.GridLine, lines []pr.GridNames, side string) (isSpan bool, _ int, _ string, coord int) {
	isSpan, number, ident := line.IsSpan(), line.Val, line.Ident
	if ident != "" && line.IsCustomIdent() {
		hasBroken := false
		var (
			line pr.GridNames
			tag  = fmt.Sprintf("{%s}-{%s}", ident, side)
		)
		for coord, line = range lines {
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
		if ident == "" {
			coord = number - 1
		} else {
			step := -1
			if number > 0 {
				step = 1
			}
			L := len(lines) / step
			hasBroken := false
			for coord := 0; coord < L; coord++ {
				line := lines[coord*step]
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
				coord += utils.Abs(number)
			}

			if step == -1 {
				coord = len(lines) - 1 - coord
			}
		}
	}
	if isSpan {
		coord = 0
	}
	return isSpan, number, ident, coord
}

type placement [2]int // coord, size

func (pl placement) isNotNone() bool { return pl[1] != 0 }

// Input coordinates are 1-indexed, returned coordinates are 0-indexed.
func getPlacement(start, end pr.GridLine, lines []pr.GridNames) placement {
	if start.Tag == pr.Auto || start.Tag == pr.Span {
		if end.Tag == pr.Auto || end.Tag == pr.Span {
			return placement{}
		}
	}
	var (
		coord, size      int
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
		coord = 0
	}
	if end.Tag != pr.Auto {
		var coordEnd int
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
				for index, line := range lines[coord+1:] {
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
		} else if coord != 0 {
			size = coordEnd - coord
		}
		if coord == 0 {
			if spanIdent == "" {
				coord = coordEnd - size
			} else {
				if number == 0 {
					number = 1
				}
				if coordEnd > 0 {
					slice := lines[coordEnd-1:]
					hasBroken := false
					for coord := range slice {
						line := slice[len(slice)-1-coord]
						if utils.IsIn(line, spanIdent) {
							number -= 1
						}
						if number == 0 {
							coord = coordEnd - 1 - coord
							hasBroken = true
							break
						}
					}
					if !hasBroken {
						coord = -number
					}
				} else {
					coord = -number
				}
			}
			size = coordEnd - coord
		}
	} else {
		size = 1
	}
	if size < 0 {
		size = -size
		coord -= size
	}
	if size == 0 {
		size = 1
	}
	return placement{coord, size}
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
	columns []pr.GridNames, childrenPositions map[Box][4]int, dense bool,
) placement {
	occupiedColumns := map[int]bool{}
	for _, rect := range childrenPositions {
		x, y, width, height := rect[0], rect[1], rect[2], rect[3]
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
			var placement placement
			if columnStart.Tag == pr.Auto {
				placement = getPlacement(pr.GridLine{Val: x + 1}, columnEnd, columns)
			} else {
				if columnStart.Tag == pr.Span {
					panic("expected span")
				}
				// If the placement contains two spans, remove the one
				// contributed by the end grid-placement property.
				// https://drafts.csswg.org/css-grid/#grid-placement-errors
				span := getSpan(columnStart)
				placement = getPlacement(columnStart, pr.GridLine{Val: x + 1 + span}, columns)
			}
			hasIntersection := false
			for col := placement[0]; col < placement[0]+placement[1]; col++ {
				if occupiedColumns[col] {
					hasIntersection = true
					break
				}
			}
			if !hasIntersection {
				return placement
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
		if columnStart.Tag == pr.Auto {
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
	sizingFunctions [][2]pr.DimOrS, tracksSizes [][2]pr.Float, span int, direction byte, containingBlock *bo.BoxFields,
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
			for _, sizes := range tracksSizes[i : i+span] {
				space -= pr.Float(sizes[affectedSizes])
			}
			space = pr.Max(0, space)
			// 2.2 Distribute space up to limits.
			var affectedTracksNumbers, unaffectedTracksNumbers []int
			for j := i; j < i+span; j++ {
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
				affectedSize := tracksSizes[trackNumber][affectedSizes]
				limit := tracksSizes[trackNumber][1]
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
					affectedSize := (tracksSizes[trackNumber][affectedSizes])
					limit := tracksSizes[trackNumber][1]
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
			tracksSizes[i][1] = tracksSizes[i][0] + increase
		} else {
			tracksSizes[i][affectedSizes] += increase
		}
	}
}

func resolveTracksSizes(context *layoutContext, sizingFunctions [][2]pr.DimOrS, boxSize, childrenPositions map[Box][4]int,
	implicitStart int, direction byte, gap,
	containingBlock *bo.BoxFields, orthogonalSizes [][2]pr.DimOrS,
) [][2]pr.Float {
	// assert direction := range "xy"
	// TODO: Check that auto box size is 0 for percentages.
	percentBoxSize := 0
	if boxSize != "auto" {
		percentBoxSize = boxSize
	}

	// 1.1 Initialize track sizes.
	tracksSizes := make([][2]pr.Float, len(sizingFunctions))
	for i, funcs := range sizingFunctions {
		minFunction, maxFunction := funcs[0], funcs[1]
		baseSize = None
		if isLength(minFunction) {
			baseSize = percentage(minFunction, percentBoxSize)
		} else if minFunction.S == "min-content" || minFunction.S == "max-content" || minFunction.S == "auto" {
			baseSize = 0
		}
		growthLimit = None
		if isLength(maxFunction) {
			growthLimit = percentage(maxFunction, percentBoxSize)
		} else if (maxFunction.S == "min-content" || maxFunction.S == "max-content" || maxFunction.S == "auto") ||
			isFr(maxFunction) {
		}
		growthLimit = inf
		if baseSize != None && growthLimit != None {
			growthLimit = max(baseSize, growthLimit)
		}
		tracksSizes[i] = [2]int{baseSize, growthLimit}
	}

	// 1.2 Resolve intrinsic track sizes.
	// 1.2.1 Shim baseline-aligned items.
	// TODO: Shim items.
	// 1.2.2 Size tracks to fit non-spanning items.
	tracksChildren := make([][]Box, len(tracksSizes))
	for child, rect := range childrenPositions {
		x, y, width, height := rect[0], rect[1], rect[2], rect[3]
		coord, size := y, height
		if direction == 'x' {
			coord, size = x, width
		}
		if size != 1 {
			continue
		}
		tracksChildren[coord-implicitStart] = append(tracksChildren[coord-implicitStart], child)
	}
	// iterable = zip(tracksChildren, sizingFunctions, tracksSizes)
	for i, children := range tracksChildren {
		minFunction, maxFunction := sizingFunctions[i][0], sizingFunctions[i][1]
		sizes := tracksSizes[i]
		if len(children) == 0 {
			continue
		}
		if direction == 'y' {
			// TODO: Find a better way to get height.
			height := 0
			for _, child := range children {
				x, _, width, _ = childrenPositions[child]
				width = sum(orthogonalSizes[x : x+width])
				child = child.deepcopy()
				child.positionX = 0
				child.positionY = 0
				parent = bo.BlockContainerBox.anonymousFrom(
					containingBlock, nil)
				resolvePercentages(parent, containingBlock)
				parent.positionX = child.positionX
				parent.positionY = child.positionY
				parent.width = width
				parent.height = height
				bottomSpace = -inf
				child, _, _, _, _, _ = blockLevelLayout(context, child, bottomSpace, nil,
					parent, true, new(), new())
				height = max(height, child.marginHeight())
			}
			if minFunction.S == "min-content" || minFunction.S == "maxContent" || minFunction.S == "auto" {
				sizes[0] = height
			}
			if maxFunction.S == "min-content" || maxFunction.S == "maxContent" {
				sizes[1] = height
			}
			if sizes[0] != nil && sizes[1] != nil {
				sizes[1] = max(sizes)
			}
			continue
		}
		if minFunction == "min-content" {
			ma := pr.Float(0)
			for _, child := range children {
				if v := minContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[0] = ma
		} else if minFunction == "max-content" {
			ma := pr.Float(0)
			for _, child := range children {
				if v := maxContentWidth(context, child, true); v > ma {
					ma = v
				}
			}
			sizes[0] = ma
		} else if minFunction == "auto" {
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
		if maxFunction == "min-content" {
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
				sizes[1] = pr.Max(sizes[0], sizes[1])
			}
		}
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
			x, y, width, height := rect[0], rect[1], rect[2], rect[3]
			coord, size := x, width
			if direction == 'y' {
				coord, size = y, height
			}
			if size != span {
				continue
			}
			hasFr := false
			for _, functions := range sizingFunctions[i : i+span+1] {
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
		distributeExtraSpace(context, sizeMin, 'i', 'm', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock)
		// 1.2.3.2 For content-based minimums.
		distributeExtraSpace(context, sizeMin, 'c', 'c', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock)
		// 1.2.3.3 For max-content minimums.
		// TODO: Respect max-content constraint.
		distributeExtraSpace(context, sizeMin, 'm', 'C', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock)
		// 1.2.3.4 Increase growth limit.
		for _, sizes := range tracksSizes {
			if sizes[0] != nil && sizes[1] != nil {
				sizes[1] = pr.Max(sizes[0], sizes[1])
			}
		}
		i = -1
		for child, rect := range childrenPositions {
			i++
			x, y, width, height := rect[0], rect[1], rect[2], rect[3]
			coord, size := x, width
			if direction == 'y' {
				coord, size = y, height
			}
			if size != span {
				continue
			}

			hasFr := false
			for _, functions := range sizingFunctions[i : i+span+1] {
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
		distributeExtraSpace(context, sizeMax, 'i', 'c', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock)
		// 1.2.3.6 For max-content maximums.
		distributeExtraSpace(context, sizeMax, 'm', 'C', tracksChildren, sizingFunctions, tracksSizes, span, direction, containingBlock)
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
	if boxSize == "auto" {
		freeSpace = None
	} else {
		sum := 0
		for _, size := range tracksSizes {
			sum += size[0]
		}

		freeSpace = boxSize - sum - (len(tracksSizes)-1)*gap
	}
	if freeSpace != nil && freeSpace > 0 {
		distributedFreeSpace = freeSpace / len(tracksSizes)
		for i, sizes := range tracksSizes {
			baseSize, growthLimit = sizes
			if baseSize+distributedFreeSpace > growthLimit {
				sizes[0] = growthLimit
				freeSpace -= growthLimit - baseSize
			} else {
				sizes[0] += distributedFreeSpace
				freeSpace -= distributedFreeSpace
			}
		}
	}
	// TODO: Respect max-width/-height.
	// 1.4 Expand flexible tracks.
	if freeSpace != nil && freeSpace <= 0 {
		// TODO: Respect min-content constraint.
		flexFraction = 0
	} else if freeSpace != nil {
		stop = false
		inflexibleTracks = set()
		for !stop {
			leftoverSpace = freeSpace
			flexFactorSum = 0
			for i, sizes := range tracksSizes {
				maxFunction := sizingFunctions[i][1]
				if isFr(maxFunction) {
					leftoverSpace += sizes[0]
					if !inflexibleTracks[i] {
						flexFactorSum += maxFunction.value
					}
				}
			}
			flexFactorSum = max(1, flexFactorSum)
			hypotheticalFrSize = leftoverSpace / flexFactorSum
			stop = true
			// iterable = enumerate(zip(tracksSizes, sizingFunctions))
			for i, sizes := range tracksSizes {
				maxFunction := sizingFunctions[i][1]
				if !inflexibleTracks[i] && isFr(maxFunction) {
					if hypotheticalFrSize*maxFunction.value < sizes[0] {
						inflexibleTracks.add(i)
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
				if maxFunction.value > 1 {
					flexFraction = max(
						flexFraction, maxFunction.value*sizes[0])
				} else {
					flexFraction = max(flexFraction, sizes[0])
				}
			}
		}
		// TODO: Respect grid items max-content contribution.
		// TODO: Respect min-* constraint.
	}
	for i, sizes := range tracksSizes {
		maxFunction := sizingFunctions[i][1]
		if isFr(maxFunction) {
			if flexFraction*maxFunction.value > sizes[0] {
				if freeSpace != nil {
					freeSpace -= flexFraction * maxFunction.value
				}
				sizes[0] = flexFraction * maxFunction.value
			}
		}
	}
	// 1.5 Expand stretched auto tracks.
	justifyContent := containingBlock.Style.GetJustifyContent()
	alignContent := containingBlock.Style.GetAlignContent()
	xStretch := direction == 'x' && justifyContent.Intersects("normal", "stretch")
	yStretch := direction == 'y' && alignContent.Intersects("normal", "stretch")
	if (xStretch || yStretch) && freeSpace != nil && freeSpace > 0 {
		var autoTracksSizes [][2]pr.Float
		for i, sizes := range tracksSizes {
			minFunction := sizingFunctions[i][0]
			if minFunction.S == "auto" {
				autoTracksSizes = append(autoTracksSizes, sizes)
			}
		}
		if len(autoTracksSizes) != 0 {
			distributedFreeSpace := freeSpace / len(autoTracksSizes)
			for i := range autoTracksSizes {
				autoTracksSizes[i][0] += distributedFreeSpace
			}
		}
	}

	return tracksSizes
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

// func gridLayout(context *layoutContext, box Box, bottomSpace, skipStack, containingBlock,
//                 pageIsEmpty, absoluteBoxes, fixedBoxes int) {

//     context.createBlockFormattingContext()

//     // Define explicit grid
// 	style := box.Box().Style
//     gridAreas := style.GetGridTemplateAreas()
//     flow := style.GetGridAutoFlow()
//     autoRows := cycle(style.GetGridAutoRows())
//     autoColumns := cycle(style.GetGridAutoColumns())
//     autoRowsBack := cycle(style.GetGridAutoRows().Reverse())
//     autoColumnsBack := cycle(style.GetGridAutoColumns().Reverse())
//     columnGap := style.GetColumnGap()
//     if columnGap.S == "normal" {
//         columnGap = 0
//     }
// 	rowGap := style.GetRowGap()
//     if rowGap == "normal" {
//         rowGap = 0
//     }

//     // TODO: Support "column" value in grid-auto-flow.
//     if   flow.Contains("column") {
//         LOGGER.warning(`"column" is ! supported in grid-auto-flow`)
//     }

//     if gridAreas.IsNone() {
//         gridAreas =  pr.GridTemplateAreas{{""}}
//     }

//     rows := getTemplateTracks(style.GetGridTemplateRows())
//     columns := getTemplateTracks(style.GetGridTemplateColumns())

//     // Adjust rows number
//     gridAreasColumns := 0
// 	if len(gridAreas) != 0 {
// 		gridAreasColumns = len(gridAreas[0])
// 	}
//     rowsDiff := (len(rows) - 1) / 2 - len(gridAreas)
//     if rowsDiff > 0 {
//         for c := 0; c < rowsDiff; c++ {
//             gridAreas = append(gridAreas, [None] * gridAreasColumns)
//         }
//     } else if rowsDiff < 0 {
//         for c := 0; c < -rowsDiff; c++ {
//             rows = append(rows, next(autoRows),nil)
//         }
//     }

//     // Adjust columns number
//     columnsDiff := (len(columns) - 1) / 2 - gridAreasColumns
//     if columnsDiff > 0 {
//         for row := range gridAreas {
//             for c := 0; c < columnsDiff; c++ {
//                 row = append(row, None)
//             }
//         }
//     } else if columnsDiff < 0 {
//         for c := 0; c < -columnsDiff; c++ {
//             columns = append(columns, next(autoColumns), nil)
//         }
//     }

//     // Add implicit line names
//     for y, row := range gridAreas {
//         for x, areaName := range row {
//             if areaName  == nil  {
//                 continue
//             }
// 			startName = areaName +  "-start"
// 			var  names []string
// 			for _, row := range extractNames(rows) {
// 				names = append(names,row...)
// 			}
//              if ! utils.IsIn(names,   startName) {
//                 rows[2*y]=  append(rows[2*y], startName)
//             }
// 				names= names[:0]
// 			for i, column := range extractNames(columns) {
// 				names = append(names,column ... )
// 			}
//              if ! utils.IsIn(names,   startName){
//                 columns[2*x] = append(columns[2*x],startName)
//             }
//         }
//     }
// 	for y := range gridAreas {
// 		row := gridAreas[len(gridAreas)-1-y] // reverse
//         for x := range row {
// 			areaName := row[len(row)-1-x] // reverse
//             if areaName  == ""  {
//                 continue
//             }
// 			endName = areaName+"-end"

// 			var  names []string
// 			for _, row := range extractNames(rows) {
// 				names = append(names,row ... )
// 			}
//              if ! utils.IsIn(names,   endName) {
//                 rows[-2*y-1]=  append(rows[-2*y-1], endName)
//             }
// 				names= names[:0]
// 			for i, column := range extractNames(columns) {
// 				names = append(names,column ... )
// 			}
//              if ! utils.IsIn(names,   endName){
//                 columns[-2*x-1] = append(columns[-2*x-1],endName)
//             }
//         }
//     }

//     // 1. Run the grid placement algorithm.

//     // 1.1 Position anything that’s not auto-positioned.
//     childrenPositions := map[Box][4]int{}
//     for _, child := range box.Box().Children {
//         columnStart := child.Box().Style.GetGridColumnStart()
//         columnEnd := child.Box().Style.GetGridColumnEnd()
//         rowStart := child.Box().Style.GetGridRowStart()
//         rowEnd := child.Box().Style.GetGridRowEnd()

//         columnPlacement := getPlacement(            columnStart, columnEnd, extractNames(columns))
//         rowPlacement := getPlacement(rowStart, rowEnd, extractNames(rows))

//         if columnPlacement.isNotNone() && rowPlacement.isNotNone() {
//             x, width = columnPlacement
//             y, height = rowPlacement
//             childrenPositions[child] = [4]int{ x, y, width, height}
//         }
// 	}

//     // 1.2 Process the items locked to a given row.
// 	children := make([]Box, len(box.Box().Children))
// 	copy(children, box.Box().Children)
// 		sort.Slice(children, func(i, j int) bool { return children[i].Box().Style.GetOrder() < children[j].Box().Style.GetOrder()})
//     for _, child := range children {
//         if _, has := childrenPositions[child]; has {
//             continue
//         }
// 		rowStart := child.style["gridRowStart"]
//         rowEnd := child.style["gridRowEnd"]
//         rowPlacement := getPlacement(rowStart, rowEnd, extractNames( rows))
//         if !rowPlacement.isNotNone() {
//             continue
//         }
// 		y, height = rowPlacement
//         columnStart = child.style["gridColumnStart"]
//         columnEnd = child.style["gridColumnEnd"]
//         x, width = getColumnPlacement(            rowPlacement, columnStart, columnEnd, columns,
//             childrenPositions,  utils.IsIn(flow, "dense"))
//         childrenPositions[child] = [4]int{x, y, width, height}
//     }

//     // 1.3 Determine the columns in range the implicit grid.
//     // 1.3.1 Start with the columns from the explicit grid.
//     implicitX1 = 0
//     implicitX2 =   0
// 	if gridAreas {
// 		implicitX2 = len(gridAreas[0])
// 	}
//     // 1.3.2 Add columns to the beginning and end of the implicit grid.
//    var remainingGridItems []Box
//     for _, child := range children {
//         if _, has :=  childrenPositions[child]; has {
//             x, _, width, _ = childrenPositions[child]
//         } else {
//             columnStart = child.style["gridColumnStart"]
//             columnEnd = child.style["gridColumnEnd"]
//             columnPlacement = getPlacement(columnStart, columnEnd, extractNames( columns))
//             remainingGridItems = append(remainingGridItems, child)
//             if columnPlacement {
//                 x, width = columnPlacement
//             } else {
//                 continue
//             }
//         }
// 		implicitX1 = utils.MinInt(x, implicitX1)
//         implicitX2 = utils.MaxInt(x + width, implicitX2)
//     }
// 	// 1.3.3 Add columns to accommodate max column span.
//     for _, child := range remainingGridItems {
//         columnStart = child.style["gridColumnStart"]
//         columnEnd = child.style["gridColumnEnd"]
//         span = 1
//         if columnStart != "auto" && columnStart[0] == "span" {
//             span = columnStart[1]
//         } else if columnEnd != "auto" && columnEnd[0] == "span" {
//             span = columnEnd[1]
//         }
// 		implicitX2 = utils.MaxInt(implicitX1 + (span || 1), implicitX2)
//     }

//     // 1.4 Position the remaining grid items.
//     implicitY1 = 0
//     implicitY2 = len(gridAreas)
//     for position := range childrenPositions.values() {
//         _, y, _, height = position
//         implicitY1 = utils.MinInt(y, implicitY1)
//         implicitY2 = utils.MaxInt(y + height, implicitY2)
//     }
// 	cursorX, cursorY = implicitX1, implicitY1
//     if utils.IsIn(flow, "dense" ) {
//         for _,child := range remainingGridItems {
//             columnStart = child.style["gridColumnStart"]
//             columnEnd = child.style["gridColumnEnd"]
//             columnPlacement = getPlacement(
//                 columnStart, columnEnd, extractNames( columns))
//             if columnPlacement {
//                 // 1. Set the row position of the cursor.
//                 cursorY = implicitY1
//                 x, width = columnPlacement
//                 cursorX = x
//                 // 2. Increment the cursor’s row position.
//                 rowStart = child.style["gridRowStart"]
//                 rowEnd = child.style["gridRowEnd"]
//                 for y := range count(cursorY) {
//                     if rowStart == "auto" {
//                         y, height = getPlacement( pr.GridLine{Val:y+1} , rowEnd, extractNames(rows))
//                     } else {
//                         // assert rowStart[0] == "span"
//                         // assert rowStart == "auto" || rowStart[0] == "span"
//                         span = getSpan(rowStart)
//                         y, height = getPlacement(
//                             rowStart, pr.GridLine{Val:y+1+span}, extractNames(rows))
//                     }
// 					 if y < cursorY {
//                         continue
//                     }
// 					hasBroken := false
// 					for row := y; row<  y + height; row++ {
//                         intersect = intersectWithChildren(x, y, width, height, childrenPositions.values())
//                         if intersect {
//                             // Child intersects with a positioned child on
//                             // current row.
// 							hasBroken = true
//                             break
//                         }
//                     }
// 					if !hasBroken{
//                         // Child doesn’t intersect with any positioned child on
//                         // any row.
//                         break
//                     }
//                 }
// 				yDiff = y + height - implicitY2
//                 if yDiff > 0 {
//                     for c := 0; c < yDiff; c++{
//                         rows = append(rows, next(autoRows), nil)
//                     }
// 					implicitY2 = y + height
//                 }
// 				// 3. Set the item’s row-start line.
//                 childrenPositions[child] = [4]int{ x, y, width, height}
//             } else {
//                 // 1. Set the cursor’s row && column positions.
//                 cursorX, cursorY = implicitX1, implicitY1
//                 for {
//                     // 2. Increment the column position of the cursor.
//                     y = cursorY
//                     rowStart = child.style["gridRowStart"]
//                     rowEnd = child.style["gridRowEnd"]
//                     columnStart = child.style["gridColumnStart"]
//                     columnEnd = child.style["gridColumnEnd"]
// 					hasBroken := false
//                     for x := cursorX; x <  implicitX2; x++ {
//                         if rowStart == "auto" {
//                             y, height = getPlacement(   pr.GridLine{Val:y + 1}, rowEnd, extractNames(rows))
//                         } else {
//                             span = getSpan(rowStart)
//                             y, height = getPlacement(   rowStart, pr.GridLine{Val:y + 1 + span},      extractNames(rows))
//                         }
// 						if columnStart == "auto" {
//                             x, width = getPlacement(pr.GridLine{Val: x + 1}, columnEnd, extractNames(columns))
//                         } else {
//                             span = getSpan(columnStart)
//                             x, width = getPlacement(   columnStart, pr.GridLine{Val:x + 1 + span}, extractNames(columns))
//                         }
// 						intersect = intersectWithChildren(                            x, y, width, height, childrenPositions.values())
//                         if intersect {
//                             // Child intersects with a positioned child.
//                             continue
//                         } else {
//                             // Free place found.
//                             // 3. Set the item’s row-/column-start lines.
//                             childrenPositions[child] =[4]int{x, y, width, height}
//                             yDiff = cursorY + height - 1 - implicitY2
//                             if yDiff > 0 {
//                                 for c := 0; c <yDiff; c++ {
//                                     rows.append(next(autoRows), nil)
//                                 }
// 								implicitY2 = cursorY + height - 1
//                             }
// 							hasBroken = true
// 							 break
//                         }
//                     }
// 					if !hasBroken {
//                         // No room found.
//                         // 2. Return to the previous step.
//                         cursorY += 1
//                         yDiff = cursorY + 1 - implicitY2
//                         if yDiff > 0 {
//                             for c := 0; c <yDiff; c++ {
//                                 rows.append(next(autoRows), nil)
//                             }
// 							implicitY2 = cursorY
//                         }
// 						cursorX = implicitX1
//                         continue
//                     }
// 					break
//                 }
//             }
//         }
//     } else {
//         for _, child := range remainingGridItems {
//             columnStart = child.style["gridColumnStart"]
//             columnEnd = child.style["gridColumnEnd"]
//             columnPlacement = getPlacement(
//                 columnStart, columnEnd, extractNames(columns))
//             if columnPlacement {
//                 // 1. Set the column position of the cursor.
//                 x, width = columnPlacement
//                 if x < cursorX {
//                     cursorY += 1
//                 }
// 				cursorX = x
//                 // 2. Increment the cursor’s row position.
//                 rowStart = child.style["gridRowStart"]
//                 rowEnd = child.style["gridRowEnd"]
//                 for cursorY := range count(cursorY) {
//                     if rowStart == "auto" {
//                         y, height = getPlacement(                           pr.GridLine{Val: cursorY + 1}, rowEnd, extractNames(rows))
//                     } else {
//                         // assert rowStart[0] == "span"
//                         // assert rowStart == "auto" || rowStart[0] == "span"
//                         span = getSpan(rowStart)
//                         y, height = getPlacement(                            rowStart,pr.GridLine{Val: cursorY + 1 + span},
//                             extractNames(rows))
//                     }
// 					if y < cursorY {
//                         continue
//                     }
// 					hasBroken := false
// 					for row := y; row < y + height;row++ {
//                         intersect = intersectWithChildren(                            x, y, width, height, childrenPositions.values())
//                         if intersect {
//                             // Child intersects with a positioned child on
//                             // current row.
// 							hasBroken = true
//                             break
//                         }
//                     }
// 					if !hasBroken {
//                         // Child doesn’t intersect with any positioned child on
//                         // any row.
//                         break
//                     }
//                 }
// 				 yDiff = y + height - implicitY2
//                 if yDiff > 0 {
//                     for  c := 0; c <yDiff; c++  {
//                         rows.append(next(autoRows), nil)
//                     }
// 					implicitY2 = y + height
//                 } // 3. Set the item’s row-start line.
//                 childrenPositions[child] = [4]int{x, y, width, height}
//             } else {
//                 for {
//                     // 1. Increment the column position of the cursor.
//                     y = cursorY
//                     rowStart = child.style["gridRowStart"]
//                     rowEnd = child.style["gridRowEnd"]
//                     columnStart = child.style["gridColumnStart"]
//                     columnEnd = child.style["gridColumnEnd"]
// 					hasBroken := false
//                     for x := cursorX; x < implicitX2; x++ {
//                         if rowStart == "auto" {
//                             y, height = getPlacement(                                pr.GridLine{Val:y + 1}, rowEnd, extractNames(rows))
//                         } else {
//                             span = getSpan(rowStart)
//                             y, height = getPlacement(                                rowStart, pr.GridLine{Val:y + 1 + span},
//                                 extractNames(rows))
//                         }
// 						 if columnStart == "auto" {
//                             x, width = getPlacement(                                 pr.GridLine{Val: x + 1}, columnEnd, extractNames(columns))
//                         } else {
//                             span = getSpan(columnStart)
//                             x, width = getPlacement(                                columnStart, pr.GridLine{Val:x + 1 + span},
//                                 extractNames(columns))
//                         }
// 						intersect = intersectWithChildren(                            x, y, width, height, childrenPositions.values())
//                         if intersect {
//                             // Child intersects with a positioned child.
//                             continue
//                         } else {
//                             // Free place found.
//                             // 2. Set the item’s row-/column-start lines.
//                             childrenPositions[child] =[4]int{x, y, width, height}
//                             hasBroken = true
// 							break
//                         }
//                     }
// 					if !hasBroken {
//                         // No room found.
//                         // 2. Return to the previous step.
//                         cursorY += 1
//                         yDiff = cursorY + 1 - implicitY2
//                         if yDiff > 0 {
//                             for c := 0; c <yDiff; c++ {
//                                 rows.append(next(autoRows), nil)
//                             }
// 							implicitY2 = cursorY
//                         }
// 						cursorX = implicitX1
//                         continue
//                     }
// 					 break
//                 }
//             }
//         }
//     }

//     for c:= 0; c <  - implicitX1; c++ {
//         columns.insert(0, next(autoColumnsBack))
//         columns.insert(0, nil)
//     }
// 	start := 0
// 	if len(gridAreas) != 0 {
// 		start = len(gridAreas[0])
// 	}
// 	for c:= start;c < implicitX2 ; c++ {
//         columns.append(next(autoColumns), nil)
//     }
// 	 for c:= 0; c <  - implicitY1; c++ {
//         rows.insert(0, next(auoRowsBack))
//         rows.insert(0, nil)
//     }
// 	for c := len(gridAreas); c < implicitY2; c++ {
//         rows.append(next(autoRows),nil)
//     }

//     // 2. Find the size of the grid container.

//     if isinstance(box, boxes.GridBox) {
//         // from .block import blockLevelWidth
//         blockLevelWidth(box, containingBlock)
//     } else {
//         // assert isinstance(box, boxes.InlineGridBox)
//         // from .inline import inlineBlockWidth
//         inlineBlockWidth(box, context, containingBlock)
//     }
// 	if box.width == "auto" {
//         // TODO: Calculate max-width.
//         box.width = containingBlock.width
//     }

//     // 3. Run the grid sizing algorithm.

//     // 3.0 List min/max sizing functions.
//     rowSizingFunctions := extractDims(rows)
//     columnSizingFunctions := extractDims( columns)

//     // 3.1 Resolve the sizes of the grid columns.
//     columnsSizes := resolveTracksSizes(        columnSizingFunctions, box.width, childrenPositions, implicitX1,
//         "x", columnGap, context, box, nil)

//     // 3.2 Resolve the sizes of the grid rows.
//     rowsSizes := resolveTracksSizes(        rowSizingFunctions, box.height, childrenPositions, implicitY1, "y",
//         rowGap, context, box,   columnsSizes)

//     // 3.3 Re-resolve the sizes of the grid columns with min-/max-content.
//     // TODO: Re-resolve.

//     // 3.4 Re-re-resolve the sizes of the grid columns with min-/max-content.
//     // TODO: Re-re-resolve.

//     // 3.5 Align the tracks within the grid container.
//     // TODO: Support safe/unsafe.
//     justifyContent := set(box.style["justifyContent"])
//     x := box.contentBoxX()
//     freeWidth = max(0, box.width - sum(size for size, _ := range columnsSizes))
//     columnsPositions = []
//     columnsNumber = len(columnsSizes)
//     if justifyContent & {"center"} {
//         x += freeWidth / 2
//         for size, _ := range columnsSizes {
//             columnsPositions.append(x)
//             x += size + columnGap
//         }
//     } else if justifyContent & {"right", "end", "flex-end"} {
//         x += freeWidth
//         for size, _ := range columnsSizes {
//             columnsPositions.append(x)
//             x += size + columnGap
//         }
//     } else if justifyContent & {"space-around"} {
//         x += freeWidth / 2 / columnsNumber
//         for size, _ := range columnsSizes {
//             columnsPositions.append(x)
//             x += size + freeWidth / columnsNumber + columnGap
//         }
//     } else if justifyContent & {"space-between"} {
//         for size, _ := range columnsSizes {
//             columnsPositions.append(x)
//             if columnsNumber >= 2 {
//                 x += size + freeWidth / (columnsNumber - 1) + columnGap
//             }
//         }
//     } else if justifyContent & {"space-evenly"} {
//         x += freeWidth / (columnsNumber + 1)
//         for size, _ := range columnsSizes {
//             columnsPositions.append(x)
//             x += size + freeWidth / (columnsNumber + 1) + columnGap
//         }
//     } else {
//         for size, _ := range columnsSizes {
//             columnsPositions.append(x)
//             x += size + columnGap
//         }
//     }

//     alignContent = set(box.style["alignContent"])
//     y = box.contentBoxY()
//     if box.height == "auto" {
//         freeHeight = 0
//     } else {
//         freeHeight = (
//             box.height -
//             sum(size for size, _ := range rowsSizes) -
//             (len(rowsSizes) - 1) * rowGap)
//         freeHeight = max(0, freeHeight)
//     }
// 	rowsPositions = []
//     rowsNumber = len(rowsSizes)
//     if alignContent & {"center"} {
//         y += freeHeight / 2
//         for size, _ := range rowsSizes {
//             rowsPositions.append(y)
//             y += size + rowGap
//         }
//     } else if alignContent & {"right", "end", "flex-end"} {
//         y += freeHeight
//         for size, _ := range rowsSizes {
//             rowsPositions.append(y)
//             y += size + rowGap
//         }
//     } else if alignContent & {"space-around"} {
//         y += freeHeight / 2 / rowsNumber
//         for size, _ := range rowsSizes {
//             rowsPositions.append(y)
//             y += size + freeHeight / rowsNumber + rowGap
//         }
//     } else if alignContent & {"space-between"} {
//         for size, _ := range rowsSizes {
//             rowsPositions.append(y)
//             if rowsNumber >= 2 {
//                 y += size + freeHeight / (rowsNumber - 1) + rowGap
//             }
//         }
//     } else if alignContent & {"space-evenly"} {
//         y += freeHeight / (rowsNumber + 1)
//         for size, _ := range rowsSizes {
//             rowsPositions.append(y)
//             y += size + freeHeight / (rowsNumber + 1) + rowGap
//         }
//     } else {
//         if alignContent & {"baseline"} {
//             // TODO: Support baseline value.
//             LOGGER.warning("Baseline alignment is not supported for grid layout")
//         } for size, _ := range rowsSizes {
//             rowsPositions.append(y)
//             y += size + rowGap
//         }
//     }

//     // 4. Lay out the grid items into their respective containing blocks.
//     // Find resumeAt row.
//     thisPageChildren = []
//     resumeRow = None
//     if skipStack {
//         skipRow = next(iter(skipStack))
//         skipHeight = (
//             sum(size for size, _ := range rowsSizes[:skipRow]) +
//             (len(rowsSizes[:skipRow]) - 1) * rowGap)
//     } else {
//         skipRow = 0
//         skipHeight = 0
//     }
// 	resumeAt = None
//     for i, rowY := range enumerate(rowsPositions[skipRow:], start=skipRow) {
//         // TODO: Check that page is ! empty.
//         if context.overflowsPage(bottomSpace, rowY - skipHeight) {
//             if i == 0 {
//                 return None, None, {"break": "any", "page": None}, [], false
//             }
// 			resumeRow = i - 1
//             resumeAt = {i-1: None}
//             for child := range children {
//                 _, y, _, _ = childrenPositions[child]
//                 if skipRow <= y <= i-2 {
//                     thisPageChildren.append(child)
//                 }
//             }
// 			 break
//         }
//     } else {
//         for child := range children {
//             _, y, _, _ = childrenPositions[child]
//             if skipRow <= y {
//                 thisPageChildren.append(child)
//             }
//         }
//     } if box.height == "auto" {
//         box.height = (
//             sum(size for size, _ := range rowsSizes[skipRow:resumeRow]) +
//             (len(rowsSizes[skipRow:resumeRow]) - 1) * rowGap)
//     }
// 	// Lay out grid items.
//     justifyItems = set(box.style["justifyItems"])
//     alignItems = set(box.style["alignItems"])
//     newChildren = []
//     baseline = None
//     nextPage = {"break": "any", "page": None}
//     // from .block import blockLevelLayout
//     for child := range thisPageChildren {
//         x, y, width, height = childrenPositions[child]
//         index = box.children.index(child)
//         if skipStack && skipStack.get(y) && index := range skipStack[y] {
//             childSkipStack = skipStack[y][index]
//         } else {
//             childSkipStack = None
//         }
// 		child = child.deepcopy()
//         child.positionX = columnsPositions[x]
//         child.positionY = rowsPositions[y] - skipHeight
//         resolvePercentages(child, box)
//         width = (
//             sum(size for size, _ := range columnsSizes[x:x+width]) +
//             (width - 1) * columnGap)
//         height = (
//             sum(size for size, _ := range rowsSizes[y:y+height]) +
//             (height - 1) * rowGap)
//         childWidth = width - (
//             child.marginLeft + child.borderLeftWidth + child.paddingLeft +
//             child.marginRight + child.borderRightWidth + child.paddingRight)
//         childHeight = height - (
//             child.marginTop + child.borderTopWidth + child.paddingTop +
//             child.marginBottom + child.borderBottomWidth + child.paddingBottom)
//     }

//         justifySelf = set(child.style["justifySelf"])
//         if justifySelf & {"auto"} {
//             justifySelf = justifyItems
//         }
// 		if justifySelf & {"normal", "stretch"} {
//             if child.style["width"] == "auto" {
//                 child.style["width"] = Dimension(childWidth, "px")
//             }
//         }
// 		alignSelf = set(child.style["alignSelf"])
//         if alignSelf & {"auto"} {
//             alignSelf = alignItems
//         }
// 		if alignSelf & {"normal", "stretch"} {
//             if child.style["height"] == "auto" {
//                 child.style["height"] = Dimension(childHeight, "px")
//             }
//         }

//         // TODO: Find a better solution for the layout.
//         parent = boxes.BlockContainerBox.anonymousFrom(box, ())
//         resolvePercentages(parent, containingBlock)
//         parent.positionX = child.positionX
//         parent.positionY = child.positionY
//         parent.width = width
//         parent.height = height
//         newChild, childResumeAt, childNextPage = blockLevelLayout(
//             context, child, bottomSpace, childSkipStack, parent,
//             pageIsEmpty, absoluteBoxes, fixedBoxes)[:3]
//         if newChild {
//             pageIsEmpty = false
//             // TODO: Support fragmentation in grid items.
//         } else {
//             // TODO: Support fragmentation in grid rows.
//             continue
//         }

//         // TODO: Apply auto margins.
//         if justifySelf & {"normal", "stretch"} {
//             newChild.width = max(childWidth, newChild.width)
//         } else {
//             newChild.width = maxContentWidth(context, newChild)
//             diff = childWidth - newChild.width
//             if justifySelf & {"center"} {
//                 newChild.translate(diff / 2, 0)
//             } else if justifySelf & {"right", "end", "flex-end", "self-end"} {
//                 newChild.translate(diff, 0)
//             }
//         }

//         // TODO: Apply auto margins.
//         if alignSelf & {"normal", "stretch"} {
//             newChild.height = max(childHeight, newChild.height)
//         } else {
//             diff = childHeight - newChild.height
//             if alignSelf & {"center"} {
//                 newChild.translate(0, diff / 2)
//             } else if alignSelf & {"end", "flex-end", "self-end"} {
//                 newChild.translate(0, diff)
//             }
//         }

//         // TODO: Take care of page fragmentation.
//         newChildren.append(newChild)
//         if baseline  == nil  && y == implicitY1 {
//             baseline = findInFlowBaseline(newChild)
//         }

//     box = box.copyWithChildren(newChildren)
//     if isinstance(box, boxes.InlineGridBox) {
//         // TODO: Synthetize a real baseline value.
//         LOGGER.warning("Inline grids are ! supported")
//         box.baseline = baseline || 0
//     }

//     context.finishBlockFormattingContext(box)

//     return box, resumeAt, nextPage, [], false
// }
