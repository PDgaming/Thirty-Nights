package cutscene

import (
	"fmt"

	"thirty.nights/src/dialogue"
)

type Session struct {
	def      *Cutscene
	step     int
	waitLeft float32
	dialogue *dialogue.Session
	finished bool
	err      error
}

func NewSession(def *Cutscene) *Session {
	s := &Session{def: def}
	if def == nil || len(def.Steps) == 0 {
		s.finished = true
		return s
	}
	s.initStep()
	return s
}

func (s *Session) Finished() bool { return s.finished }

func (s *Session) Err() error { return s.err }

// Dialogue returns the active dialogue session for the current step, if any.
func (s *Session) Dialogue() *dialogue.Session { return s.dialogue }

func (s *Session) CurrentStepType() StepType {
	if s.finished || s.def == nil || s.step < 0 || s.step >= len(s.def.Steps) {
		return ""
	}
	return s.def.Steps[s.step].Type
}

func (s *Session) Update(dt float32, dialogues dialogue.Catalog) {
	if dt < 0 {
		dt = 0
	}
	for !s.finished {
		step, ok := s.currentStep()
		if !ok {
			s.finished = true
			return
		}

		switch step.Type {
		case StepWait:
			s.waitLeft -= dt
			dt = 0
			if s.waitLeft > 0 {
				return
			}
			s.advance()
		case StepDialogue:
			if !s.ensureDialogue(step, dialogues) {
				return
			}
			s.dialogue.Update(dt)
			dt = 0
			if s.dialogue.Snapshot().Finished {
				s.dialogue = nil
				s.advance()
				continue
			}
			return
		default:
			s.fail(fmt.Errorf("cutscene: %s: step %d: unknown step type %q", s.def.ID, s.step, step.Type))
			return
		}
	}
}

// Advance forwards an interact/advance input to the current blocking step.
func (s *Session) Advance(dialogues dialogue.Catalog) {
	if s.finished {
		return
	}
	step, ok := s.currentStep()
	if !ok {
		s.finished = true
		return
	}
	if step.Type != StepDialogue {
		return
	}
	if !s.ensureDialogue(step, dialogues) {
		return
	}
	s.dialogue.Advance()
	if s.dialogue.Snapshot().Finished {
		s.dialogue = nil
		s.advance()
		// Progress through any immediately-completable following steps (e.g. zero waits not possible here,
		// but keeps state consistent). Use zero dt.
		s.Update(0, dialogues)
	}
}

func (s *Session) currentStep() (Step, bool) {
	if s.def == nil || s.step < 0 || s.step >= len(s.def.Steps) {
		return Step{}, false
	}
	return s.def.Steps[s.step], true
}

func (s *Session) ensureDialogue(step Step, dialogues dialogue.Catalog) bool {
	if s.dialogue != nil {
		return true
	}
	def, ok := dialogues.Get(step.ID)
	if !ok {
		s.fail(fmt.Errorf("cutscene: %s: step %d: unknown dialogue %q", s.def.ID, s.step, step.ID))
		return false
	}
	s.dialogue = dialogue.NewSession(def)
	return true
}

func (s *Session) initStep() {
	step, ok := s.currentStep()
	if !ok {
		s.finished = true
		return
	}
	switch step.Type {
	case StepWait:
		s.waitLeft = step.Duration
	case StepDialogue:
		s.dialogue = nil
	default:
		s.fail(fmt.Errorf("cutscene: %s: step %d: unknown step type %q", s.def.ID, s.step, step.Type))
	}
}

func (s *Session) advance() {
	s.step++
	s.waitLeft = 0
	s.dialogue = nil
	if s.def == nil || s.step >= len(s.def.Steps) {
		s.finished = true
		return
	}
	s.initStep()
}

func (s *Session) fail(err error) {
	s.err = err
	s.finished = true
	s.dialogue = nil
}
