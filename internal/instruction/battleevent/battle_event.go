// Package battleevent provides battle-event script instructions.
package battleevent

import (
	"fmt"
	"sync"

	"github.com/JackQuincy/GoWorldsCollide/internal/instruction/event"
	"github.com/JackQuincy/GoWorldsCollide/internal/memory"
)

const (
	firstCharacterAnimationSlot = 3
	lastCharacterAnimationSlot  = 6

	addRemoveEntitySourceStart = 0x01fde8
	addRemoveEntitySourceEnd   = 0x01fdec
	opcodeTableStart           = 0x01fdbe
	c1Bank                     = 0x01
)

var animationSlots = struct {
	sync.Mutex
	next byte
}{next: firstCharacterAnimationSlot}

func NOP() event.Instruction { return event.NewInstruction(0x02) }
func End() event.Instruction { return event.NewInstruction(0xff) }

func AddTarget(entityID int) event.Instruction {
	return event.NewInstruction(0x13, entityID&0x7f)
}

func RemoveTarget(entityID int) event.Instruction {
	return event.NewInstruction(0x13, entityID|0x80)
}

// AddCharacterAnimation queues a character animation in the next script slot.
func AddCharacterAnimation(character, address int) event.Instruction {
	animationSlots.Lock()
	defer animationSlots.Unlock()

	slot := animationSlots.next
	animationSlots.next++
	if animationSlots.next > lastCharacterAnimationSlot {
		panic("battleevent: too many character animations before execute")
	}

	return event.NewInstruction(
		slot,
		character,
		[]byte{byte(address), byte(address >> 8)},
	)
}

func ClearAnimations() event.Instruction { return event.NewInstruction(0x0e) }

// ExecuteAnimations executes queued animations and resets animation slot
// allocation for the next sequence.
func ExecuteAnimations() event.Instruction {
	animationSlots.Lock()
	animationSlots.next = firstCharacterAnimationSlot
	animationSlots.Unlock()
	return event.NewInstruction(0x0f)
}

func OpenMultiLineDialogWindow() event.Instruction  { return event.NewInstruction(0x11) }
func CloseMultiLineDialogWindow() event.Instruction { return event.NewInstruction(0x10) }

func DisplaySingleLineDialog(dialogID int) event.Instruction {
	return event.NewInstruction(0x00, dialogID)
}

func DisplayMultiLineDialog(dialogID int) event.Instruction {
	return event.NewInstruction(0x01, dialogID)
}

// IncrementChecksComplete increments the checks-complete event word. Call
// InstallHandlers once before writing scripts that use this instruction.
func IncrementChecksComplete() event.Instruction {
	return event.NewInstruction(0x15)
}

// HandlerAddresses records the installed C1 handlers.
type HandlerAddresses struct {
	AddRemoveEntity         int
	IncrementChecksComplete int
}

// InstallHandlers installs the custom battle-event opcode handlers. Upstream
// performs these mutations during import and first construction; Go keeps them
// explicit so ROM writes have clear ownership.
func InstallHandlers(manager *memory.Manager, checksCompleteAddress int) (HandlerAddresses, error) {
	if manager == nil {
		return HandlerAddresses{}, fmt.Errorf("battleevent: nil memory manager")
	}
	if checksCompleteAddress < 0 || checksCompleteAddress > 0xffff {
		return HandlerAddresses{}, fmt.Errorf(
			"battleevent: checks-complete address out of range: %#x",
			checksCompleteAddress,
		)
	}

	addRemoveSource, err := manager.Read(addRemoveEntitySourceStart, addRemoveEntitySourceEnd)
	if err != nil {
		return HandlerAddresses{}, fmt.Errorf("battleevent: read add/remove handler: %w", err)
	}
	addRemoveSpace, err := manager.Allocate(
		c1Bank,
		len(addRemoveSource),
		"battle event add/remove entity handler",
	)
	if err != nil {
		return HandlerAddresses{}, fmt.Errorf("battleevent: allocate add/remove handler: %w", err)
	}
	if err := addRemoveSpace.Write(addRemoveSource); err != nil {
		return HandlerAddresses{}, fmt.Errorf("battleevent: write add/remove handler: %w", err)
	}
	if err := setOpcodeAddress(manager, 0x13, addRemoveSpace.StartAddress); err != nil {
		return HandlerAddresses{}, err
	}

	incrementHandler := []byte{
		0xee,
		byte(checksCompleteAddress),
		byte(checksCompleteAddress >> 8),
		0x60,
	}
	incrementSpace, err := manager.Allocate(
		c1Bank,
		len(incrementHandler),
		"battle event increment checks complete handler",
	)
	if err != nil {
		return HandlerAddresses{}, fmt.Errorf("battleevent: allocate increment handler: %w", err)
	}
	if err := incrementSpace.Write(incrementHandler); err != nil {
		return HandlerAddresses{}, fmt.Errorf("battleevent: write increment handler: %w", err)
	}
	if err := setOpcodeAddress(manager, 0x15, incrementSpace.StartAddress); err != nil {
		return HandlerAddresses{}, err
	}

	return HandlerAddresses{
		AddRemoveEntity:         addRemoveSpace.StartAddress,
		IncrementChecksComplete: incrementSpace.StartAddress,
	}, nil
}

func setOpcodeAddress(manager *memory.Manager, opcode byte, address int) error {
	tableAddress := opcodeTableStart + int(opcode)*2
	space, err := manager.Reserve(
		tableAddress,
		tableAddress+1,
		fmt.Sprintf("battle event opcode %#x", opcode),
	)
	if err != nil {
		return fmt.Errorf("battleevent: reserve opcode %#x table entry: %w", opcode, err)
	}
	if err := space.Write(byte(address), byte(address>>8)); err != nil {
		return fmt.Errorf("battleevent: write opcode %#x table entry: %w", opcode, err)
	}
	return nil
}
