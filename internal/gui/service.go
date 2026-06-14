package gui

import "github.com/TCP404/OneTiny-cli/internal/control"

type Service struct {
	controller *control.Controller
	dialogs    DialogAdapter
}

func NewService(controller *control.Controller, dialogs DialogAdapter) *Service {
	return &Service{controller: controller, dialogs: dialogs}
}

func (s *Service) setDialogAdapter(dialogs DialogAdapter) {
	s.dialogs = dialogs
}

func (s *Service) GetStatus() (control.StatusDTO, error) {
	return s.controller.GetStatus()
}

func (s *Service) StartSharing() (control.StatusDTO, error) {
	return s.controller.StartSharing()
}

func (s *Service) StopSharing() (control.StatusDTO, error) {
	return s.controller.StopSharing()
}

func (s *Service) UpdateConfig(patch control.ConfigPatchDTO) (control.StatusDTO, error) {
	return s.controller.UpdateConfig(patch)
}

func (s *Service) SetCredentials(patch control.CredentialPatchDTO) (control.StatusDTO, error) {
	return s.controller.SetCredentials(patch)
}

func (s *Service) GetLogs(filter control.LogFilterDTO) ([]control.LogEntryDTO, error) {
	return s.controller.GetLogs(filter)
}

func (s *Service) ClearLogs() error {
	return s.controller.ClearLogs()
}

func (s *Service) ChooseDirectory(current string) (string, error) {
	if s.dialogs == nil {
		return "", nil
	}
	return s.dialogs.ChooseDirectory(current)
}

func (s *Service) ExportLogs(filter control.LogFilterDTO) (string, error) {
	if s.dialogs == nil {
		return "", nil
	}
	path, err := s.dialogs.ChooseExportPath()
	if err != nil || path == "" {
		return path, err
	}
	if err := s.controller.ExportLogs(path, filter); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) OpenConfigDir() error {
	if s.dialogs == nil {
		return nil
	}
	return s.dialogs.OpenConfigDir()
}

func (s *Service) requestQuit(onConfirmed func()) bool {
	status, err := s.controller.GetStatus()
	if err != nil {
		return false
	}
	if !status.Running {
		return true
	}
	if s.dialogs == nil {
		return false
	}
	err = s.dialogs.ConfirmQuitWhileRunning(func() {
		s.shutdown()
		if onConfirmed != nil {
			onConfirmed()
		}
	})
	if err != nil {
		return false
	}
	return false
}

func (s *Service) shutdown() {
	_, _ = s.controller.StopSharing()
}
