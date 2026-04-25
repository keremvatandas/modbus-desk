package main

import (
	"context"

	"modbusdesk/internal/modbus"
	"modbusdesk/internal/profiles"
)

// App struct
type App struct {
	ctx      context.Context
	modbus   *modbus.Service
	profiles *profiles.Store
}

// NewApp creates a new App application struct
func NewApp() *App {
	profilePath, err := profiles.DefaultPath()
	if err != nil {
		profilePath = "profiles.json"
	}
	return &App{
		modbus:   modbus.NewService(),
		profiles: profiles.NewStore(profilePath),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Connect(req modbus.ConnectRequest) (modbus.ConnectionStatus, error) {
	return a.modbus.Connect(req)
}

func (a *App) Disconnect() (modbus.ConnectionStatus, error) {
	return a.modbus.Disconnect(), nil
}

func (a *App) GetConnectionStatus() modbus.ConnectionStatus {
	return a.modbus.Status()
}

func (a *App) ReadRegisters(req modbus.ReadRegistersRequest) (modbus.ReadRegistersResponse, error) {
	return a.modbus.ReadRegisters(req)
}

func (a *App) ReadBits(req modbus.ReadBitsRequest) (modbus.ReadBitsResponse, error) {
	return a.modbus.ReadBits(req)
}

func (a *App) PreviewWriteSingleCoil(req modbus.WriteSingleCoilRequest) (modbus.WritePreview, error) {
	return a.modbus.PreviewWriteSingleCoil(req)
}

func (a *App) WriteSingleCoil(req modbus.WriteSingleCoilRequest) error {
	return a.modbus.WriteSingleCoil(req)
}

func (a *App) PreviewWriteSingleRegister(req modbus.WriteSingleRegisterRequest) (modbus.WritePreview, error) {
	return a.modbus.PreviewWriteSingleRegister(req)
}

func (a *App) WriteSingleRegister(req modbus.WriteSingleRegisterRequest) error {
	return a.modbus.WriteSingleRegister(req)
}

func (a *App) PreviewWriteMultipleCoils(req modbus.WriteMultipleCoilsRequest) (modbus.WritePreview, error) {
	return a.modbus.PreviewWriteMultipleCoils(req)
}

func (a *App) WriteMultipleCoils(req modbus.WriteMultipleCoilsRequest) error {
	return a.modbus.WriteMultipleCoils(req)
}

func (a *App) PreviewWriteMultipleRegisters(req modbus.WriteMultipleRegistersRequest) (modbus.WritePreview, error) {
	return a.modbus.PreviewWriteMultipleRegisters(req)
}

func (a *App) WriteMultipleRegisters(req modbus.WriteMultipleRegistersRequest) error {
	return a.modbus.WriteMultipleRegisters(req)
}

func (a *App) GetFrameLogs() []modbus.FrameLog {
	return a.modbus.FrameLogs()
}

func (a *App) ListProfiles() ([]profiles.Profile, error) {
	return a.profiles.List()
}

func (a *App) SaveProfile(profile profiles.Profile) error {
	return a.profiles.Save(profile)
}

func (a *App) DeleteProfile(name string) error {
	return a.profiles.Delete(name)
}
