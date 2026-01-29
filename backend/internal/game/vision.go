package game

// CalculateEffectiveVisionRadius returns the effective vision radius for an agent,
// including base vision, upgrade bonuses, and beacon bonuses from owned structures.
func CalculateEffectiveVisionRadius(agent *Agent, baseRadius int, worldObjects *WorldObjectManager) int {
	pos := agent.GetPosition()
	visionRadius := agent.GetEffectiveVision(baseRadius)

	beacons := worldObjects.GetByOwner(agent.ID)
	for _, beacon := range beacons {
		if beacon.Type == ObjectStructure && beacon.StructureType == StructureBeacon {
			dx := beacon.Position.X - pos.X
			dy := beacon.Position.Y - pos.Y
			if dx >= -visionRadius && dx <= visionRadius && dy >= -visionRadius && dy <= visionRadius {
				visionRadius += beacon.VisionBonus
			}
		}
	}

	return visionRadius
}
