package level

import (
	"github.com/Relixik/minecraft-server/apis/buff"
	"github.com/Relixik/minecraft-server/apis/data/tags"
	apis_level "github.com/Relixik/minecraft-server/apis/game/level"
	"github.com/Relixik/minecraft-server/impl/base"
)

type chunk struct {
	x int
	z int

	level *level

	slices []*slice

	heightMap map[heightMapType]*heightMap
}

func newChunk(level *level, x, z int) *chunk {
	chunk := &chunk{
		x: x,
		z: z,

		level: level,

		slices:    make([]*slice, apis_level.SliceC, apis_level.SliceC),
		heightMap: make(map[heightMapType]*heightMap),
	}

	for _, mapType := range heightMapTypes {
		chunk.heightMap[mapType] = &heightMap{
			chunk: chunk,

			heightMapType: mapType,
			heightMapData: base.NewCompacter(9, 256),
		}
	}

	return chunk
}

func (c *chunk) ChunkX() int {
	return c.x
}

func (c *chunk) ChunkZ() int {
	return c.z
}

func (c *chunk) Level() apis_level.Level {
	return c.level
}

func (c *chunk) Slices() []apis_level.Slice {
	slices := make([]apis_level.Slice, apis_level.SliceC, apis_level.SliceC)

	for index, slice := range c.slices {
		slices[index] = slice
	}

	return slices
}

func (c *chunk) GetSlice(y int) apis_level.Slice {
	// Validate slice index and return a safe default slice for invalid indices
	if y < 0 || y > 15 {
		// Return bedrock slice (index 0) as a safe fallback
		if c.slices[0] == nil {
			c.slices[0] = newSlice(c, 0)
		}
		return c.slices[0]
	}

	slc := c.slices[y]
	if slc != nil {
		return slc
	}

	gen := newSlice(c, y)
	c.slices[y] = gen

	return gen
}

func (c *chunk) GetBlock(x, y, z int) apis_level.Block {
	// Validate coordinates and return air block for invalid coordinates
	if x < 0 || x > 15 || y < 0 || y > 255 || z < 0 || z > 15 {
		// Return a safe air block at chunk origin instead of panicking
		bedrock := c.GetSlice(0).(*slice) // get bedrock slice with type assertion
		return &block{
			x:     c.x << 0x04, // chunk origin x
			y:     0,           // bedrock level (safe y)
			z:     c.z << 0x04, // chunk origin z
			slice: bedrock,
		}
	}

	return &block{
		x: (c.x << 0x04) | x,
		y: y,
		z: (c.z << 0x04) | z,

		slice: c.GetSlice(y >> 0x04).(*slice),
	}
}

func (c *chunk) Push(writer buff.Buffer) {
	mask := int32(0)

	for i := 0; i < apis_level.SliceC; i++ {
		if len(c.slices) < i {
			break
		}

		mask |= 1 << i

		slice := c.slices[i]
		slice.Push(writer)
	}

	writer.PushVrI(mask)
}

func (c *chunk) HeightMapNbtCompound() *tags.NbtCompound {
	compound := tags.NbtCompound{Value: make(map[string]tags.Nbt)}

	motionBlocking := c.heightMap[MotionBlocking]
	compound.Set(string(motionBlocking.heightMapType), &tags.NbtArrI64{Value: motionBlocking.heightMapData.Values})

	return &compound
}

type heightMapType string

const (
	WorldSurfaceWg         heightMapType = "WORLD_SURFACE_WG"
	WorldSurface           heightMapType = "WORLD_SURFACE"
	OceanFloorWg           heightMapType = "OCEAN_FLOOR_WG"
	OceanFloor             heightMapType = "OCEAN_FLOOR"
	MotionBlocking         heightMapType = "MOTION_BLOCKING"
	MotionBlockingNoLeaves heightMapType = "MOTION_BLOCKING_NO_LEAVES"
)

var heightMapTypes = []heightMapType{
	WorldSurfaceWg,
	WorldSurface,
	OceanFloorWg,
	OceanFloor,
	MotionBlocking,
	MotionBlockingNoLeaves,
}

type heightMap struct {
	chunk *chunk
	check func(b *block) bool

	heightMapType heightMapType
	heightMapData *base.Compacter
}
