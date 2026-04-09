# Polaris .Program.polaris - Format Reference

## Document Structure

```
//FILE_VALIDATION=<sha256_hex>     ← SHA-256 hash of JSON content below
[
  MaterialSection,                  ← Material properties (density, type)
  PieceSection,                     ← Physical piece geometry & operations
  ProgramSection                    ← CNC machine configuration & production params
]
```

The `FILE_VALIDATION` hash detects unauthorized modifications. Computed from JSON bytes (excluding the first line). Warning-only on mismatch - file still loads.

---

## 1. MaterialObject

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **Code** | string | `"S355ML"` | Material grade identifier |
| **SpecificWeight** | float | `7.85` | Density in g/cm³ (steel = 7.85). Used for weight calculations |
| **MaterialTypeAttribute** | string | `"MildSteel"` | Material classification for manufacturing rules |

---

## 2. PieceObject

### Identification Fields

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **Contract** | string | `"B2149"` | Contract/order number |
| **Project** | string | `"B2149"` | Project identifier |
| **Drawing** | string | `"MONOPILE_PARTS_A"` | Engineering drawing reference |
| **Assembly** | string | `""` | Assembly identifier (empty for standalone parts) |
| **Part** | string | `"POS_296_R01"` | Part number. J-BEVEL adds suffix: `POS_296_R_J-BEVEL-2` |
| **ProfileType** | string | `"P"` | Profile category ("P" = plate) |

### PieceAttributes

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **ExecutedQuantity** | int | `0` | Pieces already manufactured |
| **TotalQuantity** | int | `1` | Total pieces to manufacture |
| **Length** | float | `29591.76` | X-axis dimension in mm |
| **Width** | float | `3681.0` | Y-axis dimension in mm |
| **Thickness** | float | `80.0` | Material thickness in mm |
| **MaterialCode** | string | `"S355ML"` | Links to MaterialObject.Code |
| **GrossWeight** | float | `68406.33` | Bounding rectangle weight in kg = `L × W × T × density / 1e6` |
| **NetWeight** | float | `68312.13` | Actual piece weight in kg. CAN ≈ Gross; CONE < Gross (banana area) |
| **IsPieceToMeasure** | bool | `true` | Custom/prototype piece (true for generated programs) |
| **IsFrequentlyManufactured** | bool | `false` | Standard recurring piece |
| **Notes** | string | `""` | Additional notes |

---

## 3. SVGEntity (JSON string embedded in JSON)

Encodes 2D contour geometry for visualization and CNC path planning.

```json
{
  "plateContourSections": [...],   ← Bounding rectangle (always M→L→L→L→Z)
  "nestedParts": [...]              ← 4 parts (P1-P4) defining actual cut edges
}
```

### Path Operations

| Operation | Fields | Description |
|-----------|--------|-------------|
| **M** | x, y | Move to point (start new path) |
| **L** | x, y | Line to point (straight cut) |
| **A** | x, y, rx, ry, thetax, largeArcFlag, sweepFlag | Arc to point (curved cut) |
| **Z** | — | Close path (return to start) |

### Arc Parameters (A operation)

| Field | Description |
|-------|-------------|
| **rx, ry** | Arc radii (equal for circular arcs, e.g. `72680.0`) |
| **thetax** | Arc rotation angle (typically `0.0`) |
| **largeArcFlag** | `0` = minor arc, `1` = major arc |
| **sweepFlag** | `0` = counterclockwise (bottom/concave), `1` = clockwise (top/convex) |

### CAN Nested Parts (rectangle)

```
P1: Left edge      (0, W) → (0, 0)           Line
P2: Bottom edge     (0, 0) → (L, 0)           Line
P3: Right edge      (L, 0) → (L, W)           Line
P4: Top edge        (L, W) → (0, W)           Line
```

### CONE Nested Parts (banana)

```
P1: Left diagonal   (0, H) → (offset, 0)      Line
P2: Right diagonal   (L-offset, 0) → (L, H)    Line
P3: Bottom arc       (offset, 0) → (L-offset, 0)  Arc (sweepFlag=0, concave)
P4: Top arc          (L, H) → (0, H)              Arc (sweepFlag=1, convex)
```

Where: L=Length, W=Width, H=PieceHeight, offset=LeftOffset

### NestedPart Fields

| Field | Type | Description |
|-------|------|-------------|
| **x, y** | float | Position offset (always `0.001`) |
| **part.name** | string | Part ID: `"P1"`, `"P2"`, `"P3"`, `"P4"` |
| **part.index** | int | Sequential index (1, 2, 3, 4) |
| **part.segments** | array | SVG path operations for this edge |
| **part.innnerSegments** | array | Inner edges/holes (typically empty). Note: `innner` is intentional - Polaris uses this spelling |

---

## 4. Operations Tree

Hierarchical manufacturing steps. Flat list with parent references.

### Tree Structure

```
Line 0: ROOT NODE          (Level 0, Parent 0)
  Line 1: FixGroup 0 NODE  (Level 1, Parent 0)
    Line 2: PATHM P1        (Level 2, Parent 1)  ← left edge
    Line 3: PATHM P2/P3     (Level 2, Parent 1)  ← bottom edge/arc
    Line 4: PATHM P3/P2     (Level 2, Parent 1)  ← right edge
    Line 5: PATHM P4        (Level 2, Parent 1)  ← top edge/arc
```

### Operation Fields

| Field | Type | Description |
|-------|------|-------------|
| **OperationType** | string | `"NODE"` (group) or `"PATHM"` (cut path) |
| **LineNumber** | int | Unique sequence number (0, 1, 2...) |
| **Level** | int | Hierarchy depth (0=root, 1=group, 2=paths) |
| **ParentLineNumber** | int | LineNumber of parent operation |
| **ToBeSkipped** | bool | Whether to skip this operation |
| **Attributes** | object | NODE or PATHM attributes (see below) |
| **AdditionalItems** | array | PATHVERTEX points (only for PATHM) |

### NODE Attributes

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **Name** | string | `"Root"` / `"FixGroup 0"` | Node identifier |
| **SequenceIndex** | int | `0` / `1` | Position in sequence |
| **OrderType** | string | `"NotDefined"` / `"SingleSide"` | Workpiece clamping mode. SingleSide = clamped on one side |
| **Ox, Oy** | float | `0.0` | Origin offset X/Y |
| **TypeOX, TypeOY** | string | `"Standard"` | Origin reference type |
| **NodeCustomType** | int | `0` / `1` | 0 = Root node, 1 = FixGroup node |
| **ProbeCode** | string | `"0"` | Probe/sensor ID ("0" = no probing) |
| **ProbeAllAoperations** | bool | `false` | Probe every operation |
| **ProbeReference** | string | `"None"` | Probing alignment reference |

### PATHM Attributes (Cutting Path)

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **Side** | string | `"C"` | Cut side: **C**=Center (no offset), L=Left, R=Right |
| **CompensationType** | string | `"None"` | Tool offset compensation mode |
| **Depth** | float | `0.0` | Cutting depth in mm (0.0 = through-cut) |
| **PitchStart** | float | `0.0` | Initial pitch distance |
| **PitchDepth** | float | `0.0` | Pitch change per cycle |
| **HeadCycle** | string | `"SingleHead"` | Cutting head mode (SingleHead / DoubleHead) |
| **DeltaY** | float | `0.0` | Y-axis offset adjustment |
| **TS** | string | `"POCKET_75"` | **Tool Specification**: POCKET = plasma cutting mode, 75 = nozzle area in mm². Selects pre-configured cutting parameters (current, gas, speed) |
| **BevelTopHeight** | float | `0.0` | Top bevel height in mm (0 = no bevel) |
| **BevelTopAngle** | float | `0.0` | Top bevel angle in degrees |
| **BevelBottomHeight** | float | `0.0` | Bottom bevel height in mm |
| **BevelBottomAngle** | float | `0.0` | Bottom bevel angle in degrees |
| **DN** | float | `0.0` | Tool diameter / offset distance |
| **COD** | string | `""` | Code of Designation for special operations |
| **BevelPathEnumeration** | int | `-1` / `1` | **Cut direction**: **-1** = reverse (left/upward edges), **1** = forward (right/bottom edges). Alternates -1,1,-1,1 for the 4 edges. Controls tool approach, chip flow, and collision avoidance |
| **XOut, YOut** | float | `0.0` | Exit/retract coordinates |
| **ProbeTS** | string | `"NotDefined"` | Probe tool specification |
| **ProbeDN** | float | `0.0` | Probe offset distance |
| **ProbeCOD** | string | `""` | Probe code designation |
| **AuxiliariesCycle** | string | `"Active"` | Auxiliary systems (cooling, gas flow): Active / Inactive |
| **TangentialSpeedOverride** | float | `0.0` | Speed override (0.0 = use default) |
| **CounteractingCycle** | bool | `false` | Anti-vibration mode |
| **ProbeCode** | string | `"0"` | Probe identifier |
| **ProbeReference** | string | `"None"` | Probe reference point |

### PATHVERTEX Attributes (Cut Points)

Each PATHM has 2 PATHVERTEX items: start point and end point.

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **X** | float | `29591.76` | X coordinate in mm |
| **Y** | float | `3681.0` | Y coordinate in mm |
| **Radius** | float | `0.0` / `-72680.0` / `75870.54` | **Arc radius**: `0.0` = straight line; **negative** = concave arc (bottom, material removed); **positive** = convex arc (top, material added) |
| **InterpolationType** | string | `"None"` | Interpolation method (None = linear) |
| **TangentialSpeedOverride** | float | `0.0` | Per-vertex speed override |

---

## 5. ProgramObject

### Core Fields

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **Name** | string | `"PGM_B2149_POS_296_R01_A"` | Program ID: `PGM_{Contract}_{Part}_{WorkingArea}` |
| **ProgramType** | string | `"PieceToMeasure"` | Job type (custom/prototype) |
| **SVGEntity** | string | (JSON) / `{"error":"unknown profile type"}` | CAN: copy of piece SVG. CONE: error placeholder |
| **Stock** | null | `null` | Raw material stock reference |
| **IsoCodeLines** | array | `[]` | ISO G-code output (empty for definition files) |

### ProgramAttributes

#### Quantity & Status
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **TotalQuantity** | int | `1` | Pieces to produce |
| **ExecutedQuantity** | int | `0` | Already produced |
| **ProfileType** | string | `"P"` | Profile category (P = plate) |
| **Length** | float | `29591.76` | Program length in mm (= Piece.Length) |

#### Machine Configuration
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **WorkingArea** | string | `"A"` / `"AB"` | Machine work zone. A = zone A, AB = both zones |
| **PlateXPosition** | float | `5220.0` | Initial X position of plate on machine bed in mm |
| **CarriageType** | string | `"VerticalClamp"` | Material transport/clamping mechanism |
| **PincherAPosition** | float | `0.0` | Clamp A position in mm |
| **PincherBPosition** | float | `0.0` | Clamp B position in mm |
| **PincherCPosition** | float | `0.0` | Clamp C position in mm |
| **RequestedStationNumber** | int | `1` | Production station number |
| **PuFlangesPosition** | string | `"FlangesDownward"` | Flange orientation for pickup/unload |

#### Cutting & Plasma
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **PlasmaTechnology** | string | `"NotSpecified"` | Plasma cutting technology level |
| **ProgramPlasmaCurrent** | float | `0.0` | Plasma arc current in amps (0 = auto) |
| **NozzleShape** | string | `"Bevel"` | Plasma nozzle geometry (Bevel / Straight / Conical) |
| **PlasmaGasCut** | string | `"NotSelected"` | Cutting gas type |
| **ShieldGasCut** | string | `"NotSelected"` | Shielding gas type |
| **GasMarking** | string | `"NotSelected"` | Gas for marking operations |
| **ProbingDistance** | float | `0.0` | Probe distance for arc initiation in mm |
| **DrillsCoolingLimit** | float | `0.0` | Temperature limit for drill cooling in °C |

#### Processing Modes
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **ProcessingModality** | string | `"0"` | Processing mode (0 = standard, 1/2 = multi-pass) |
| **DrillingToolSelection** | string | `"ByMaxLife"` | Tool selection: ByMaxLife (cost opt.) / ByQualityIndex (quality opt.) |
| **MultipleTool** | bool | `false` | Use multiple tools in sequence |
| **FinalCutModality** | bool | `false` | Final edge finishing cut |
| **SimulatedCuts** | bool | `false` | Dry-run simulation mode |
| **AutomaticReprocessing** | bool | `false` | Auto re-cut failed edges |
| **AutomaticPiecesQuantityProcessing** | bool | `false` | Auto-adjust piece count |
| **DoublePassSlantedCutsFlanges** | bool | `false` | Two-pass cutting for flanges |
| **DoublePassTubeProfile** | bool | `false` | Two-pass for tube sections |
| **TransverseSectionsProfile** | bool | `false` | Special transverse section processing |
| **CounteractingCycle** | (via PATHM) | | Anti-vibration mode |

#### Cut Strategy
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **CutsType** | string | `"NotDefined"` | Cut separation strategy |
| **SeparationCutsType** | string | `"NotDefined"` | How pieces separate from stock |
| **SurplusType** | string | `"FinalEdge"` | Where to keep surplus: FinalEdge / LeadingEdge / Balanced |
| **SurplusEnableStock** | bool | `false` | Track surplus in stock pool |
| **SurplusMinLengthStock** | float | `0.0` | Min surplus stock length in mm |
| **PreHolesOrderModality** | string | `"GroupAhead"` | Drill order: GroupAhead (all holes first) / Sequential |
| **ChipRemovalModality** | string | `"None"` | Chip evacuation: None / Blow / Suction |
| **HolesNumberBrushing** | int | `0` | Holes needing brush/reaming finishing |

#### Scrap & Waste
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **NestingLength** | float | `0.0` | Nested pattern length (0 = single piece) |
| **NestingScrap** | float | `0.0` | Scrap between nested pieces |
| **FreeStock** | float | `0.0` | Free stock/lead-in distance in mm |
| **LeadingEdgeScrap** | float | `0.0` | Scrap at leading edge |
| **FinalEdgeScrap** | float | `0.0` | Scrap at trailing edge |
| **CutScrap** | float | `0.0` | Scrap from cutting process |
| **SawingRotationScrap** | float | `0.0` | Scrap from rotation operations |
| **MaxPercentageScrap** | float | `0.0` | Max allowable scrap % (0 = unlimited) |
| **UnloadingWidth** | float | `0.0` | Unloading station width |

#### Marking
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **ScribingMarkingType** | string | `"Standard"` | Scribing method (Standard / Laser / Ink) |
| **PlasmaMarkingType** | string | `"Standard"` | Plasma marking style |
| **SurplusMarkingText1-4** | string | `""` | Text labels for surplus pieces (up to 4 lines) |

#### Production Info
| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **DestinationCode** | int | `0` | Target work center/location |
| **BatchNumber** | int | `0` | Current batch number |
| **BatchTotalNumber** | int | `0` | Total batches in order |
| **PackageNumber** | int | `0` | Package/container number |
| **SandblustingCode** | int | `0` | Sandblasting specification (0 = none) |
| **CustomerName** | string | `"SIF"` | Customer identifier |
| **PLMVersion** | string | `"2025 build 33"` | Software version that created the program |
| **MinRequiredConsole** | string | `"NotSpecified"` | Minimum console version required |
| **RequestedHeatNumber** | string | `""` | Required material heat/batch |
| **RequestedSupplier** | string | `""` | Preferred material supplier |
| **ProfileCode** | string | `""` | Profile template ID |
| **MaterialCode** | string | `""` | Material reference (in ProgramAttributes) |
| **Notes** | string | `""` | Program notes |

---

## 6. RelatedPieces

Links program to piece definition with nesting/layout parameters.

| Field | Type | Example | Description |
|-------|------|---------|-------------|
| **Contract/Project/Drawing/Assembly/Part** | string | | Must match PieceObject fields |
| **GroupId** | int | `1` | Nesting group number |
| **OffsetX, OffsetY** | float | `0.0` | Piece position offset in mm |
| **OffsetAngle** | float | `0.0` | Rotation angle in degrees |
| **RepetitionsNumber** | int | `1` | How many copies of this piece |
| **LongitudinalRotation** | bool | `false` | 180° flip along length axis |
| **TransverseRotation** | bool | `false` | 180° flip along width axis |
| **InitialDelta** | float | `0.0` | Distance from leading edge |
| **FinalDelta** | float | `0.0` | Distance from trailing edge |
| **UnloadingCodeArea** | int | `0` | Unloading destination zone |
| **Lot** | string | `""` | Lot identifier for tracking |
| **LotNumber / LotTotalNumber** | int | `0` | Lot sequence / total |
| **ExtraProcessingCode** | int | `0` | Additional processing (0 = none) |
| **GeneratedBarName** | string | `""` | Auto-generated barcode/label |
| **DestinationCode** | int | `0` | Final destination code |

---

## 7. CAN vs CONE - Key Differences

| Aspect | CAN (rectangle) | CONE (banana) |
|--------|-----------------|---------------|
| **Shape** | 4 straight edges | 2 diagonal lines + 2 arcs |
| **SVG segments** | All `L` (lines) | `L` + `A` (arcs) for P3/P4 |
| **PATHVERTEX Radius** | Always `0.0` | Bottom: `-72680.0` (concave), Top: `+75870.54` (convex) |
| **Radius sign** | — | Negative = concave (material removed), Positive = convex (material included) |
| **ProgramObject SVGEntity** | Copy of piece SVG | `{"error":"unknown profile type"}` |
| **NetWeight** | = GrossWeight | < GrossWeight (banana area < rectangle) |
| **Deviation method** | Direct coordinate replacement | Proportional scaling |
| **Extra params** | — | PieceHeight, LeftOffset, BottomRadius, TopRadius |

---

## 8. Weight Formulas

**GrossWeight** (both CAN and CONE):
```
GrossWeight = Length × Width × Thickness × SpecificWeight / 1,000,000
              (mm)    (mm)    (mm)         (g/cm³)          → kg
```

**NetWeight CAN**: = GrossWeight

**NetWeight CONE**:
```
bottomChord = Length - 2 × LeftOffset
topChord    = Length
trapArea    = (bottomChord + topChord) / 2 × PieceHeight
bottomSeg   = circularSegmentArea(bottomChord, BottomRadius)
topSeg      = circularSegmentArea(topChord, TopRadius)
netArea     = trapArea - bottomSeg + topSeg
NetWeight   = netArea × Thickness × SpecificWeight / 1,000,000
```

---

## 9. File Naming

```
PGM_{Contract}_{Part}_{WorkingArea}.Program.polaris
```

Examples:
- `PGM_B2149_POS_296_R01_A.Program.polaris`
- `PGM_B2149_POS_186_C02_A.Program.polaris`
- `PGM_B2149_POS_296_R01_J-BEVEL-2_AB.Program.polaris`

Deviation output adds suffix:
- `PGM_B2149_POS_296_R01_A_DEV_L-25_W-30.Program.polaris`
