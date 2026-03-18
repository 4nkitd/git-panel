package app

// HitZone identifies a clickable region in the UI.
type HitZone int

const (
	ZoneNone HitZone = iota
	ZoneTitle
	ZoneCommitInput
	ZoneCommitButton
	ZoneStagedHeader
	ZoneStagedFile    // + file index in zoneIndex
	ZoneUnstagedHeader
	ZoneUnstagedFile  // + file index in zoneIndex
	ZoneStashHeader
	ZoneStashFile
	ZoneDiff
	ZoneGraphHeader
	ZoneGraphCommit
	ZoneGraphFile
	ZoneStatusBar
	ZoneHelpBar
)

// RowHit maps a terminal row to a clickable zone.
type RowHit struct {
	Zone      HitZone
	FileIndex int    // index within section's file list (-1 if not a file row)
	ColAction int    // column where action icons start (-1 if none)
}

// LayoutMap tracks which row corresponds to which UI element.
// Built during View(), consumed during Update() for mouse hits.
type LayoutMap struct {
	Rows []RowHit
}

func NewLayoutMap(height int) *LayoutMap {
	rows := make([]RowHit, height)
	for i := range rows {
		rows[i] = RowHit{Zone: ZoneNone, FileIndex: -1, ColAction: -1}
	}
	return &LayoutMap{Rows: rows}
}

func (lm *LayoutMap) Set(row int, zone HitZone, fileIndex int) {
	if row >= 0 && row < len(lm.Rows) {
		lm.Rows[row] = RowHit{Zone: zone, FileIndex: fileIndex, ColAction: -1}
	}
}

func (lm *LayoutMap) SetWithAction(row int, zone HitZone, fileIndex int, colAction int) {
	if row >= 0 && row < len(lm.Rows) {
		lm.Rows[row] = RowHit{Zone: zone, FileIndex: fileIndex, ColAction: colAction}
	}
}

func (lm *LayoutMap) Get(row int) RowHit {
	if row >= 0 && row < len(lm.Rows) {
		return lm.Rows[row]
	}
	return RowHit{Zone: ZoneNone, FileIndex: -1, ColAction: -1}
}
