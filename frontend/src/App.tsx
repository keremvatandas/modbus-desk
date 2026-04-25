import {useEffect, useMemo, useState} from 'react';
import './App.css';
import {
    Connect,
    DeleteProfile,
    Disconnect,
    GetConnectionStatus,
    GetFrameLogs,
    ListProfiles,
    PreviewWriteMultipleCoils,
    PreviewWriteMultipleRegisters,
    PreviewWriteSingleCoil,
    PreviewWriteSingleRegister,
    ReadBits,
    ReadRegisters,
    SaveProfile,
    WriteMultipleCoils,
    WriteMultipleRegisters,
    WriteSingleCoil,
    WriteSingleRegister,
} from '../wailsjs/go/main/App';
import {modbus, profiles} from '../wailsjs/go/models';
import {BitTable} from './components/BitTable';
import {ConnectionPanel} from './components/ConnectionPanel';
import {FrameInspector} from './components/FrameInspector';
import {ProfilesPanel} from './components/Profiles';
import {ReadPanel} from './components/ReadPanel';
import {RegisterTable} from './components/RegisterTable';
import {WritePanel, type WriteFormState} from './components/WritePanel';

const defaultConnection: modbus.ConnectRequest = {
    host: '127.0.0.1',
    port: 502,
    unitId: 1,
    timeoutMs: 3000,
    retries: 1,
};

const defaultStatus: modbus.ConnectionStatus = {
    connected: false,
    host: '',
    port: 502,
    unitId: 1,
    timeoutMs: 3000,
    retries: 1,
    lastError: '',
    lastResponseMs: 0,
};

const defaultRead: modbus.ReadRegistersRequest = {
    functionCode: 3,
    address: 0,
    quantity: 2,
    dataType: 'uint16',
    byteOrder: 'ABCD',
    addressBase: 0,
};

const defaultWrite: WriteFormState = {
    functionCode: 6,
    address: 0,
    coilValue: false,
    registerValue: 0,
    coilValues: '1, 0, 1, 1',
    registerValues: '230, 0x00E6',
    dataType: 'uint16',
    byteOrder: 'ABCD',
};

type BusyState = 'connect' | 'read' | 'write' | 'profile' | '';
type ReadResultKind = 'registers' | 'bits';

function App() {
    const [connection, setConnection] = useState<modbus.ConnectRequest>(defaultConnection);
    const [status, setStatus] = useState<modbus.ConnectionStatus>(defaultStatus);
    const [readRequest, setReadRequest] = useState<modbus.ReadRegistersRequest>(defaultRead);
    const [writeRequest, setWriteRequest] = useState<WriteFormState>(defaultWrite);
    const [writePreview, setWritePreview] = useState<modbus.WritePreview | null>(null);
    const [rows, setRows] = useState<modbus.RegisterRow[]>([]);
    const [bitRows, setBitRows] = useState<modbus.BitRow[]>([]);
    const [readResultKind, setReadResultKind] = useState<ReadResultKind>('registers');
    const [frameLogs, setFrameLogs] = useState<modbus.FrameLog[]>([]);
    const [profileList, setProfileList] = useState<profiles.Profile[]>([]);
    const [profileName, setProfileName] = useState('Field Device');
    const [busy, setBusy] = useState<BusyState>('');
    const [notice, setNotice] = useState('Ready');
    const [error, setError] = useState('');

    useEffect(() => {
        void refreshAppState();
    }, []);

    const endpoint = useMemo(() => {
        const host = status.host || connection.host;
        const port = status.port || connection.port;
        return `${host}:${port}`;
    }, [connection.host, connection.port, status.host, status.port]);

    async function refreshAppState() {
        try {
            const [nextStatus, nextProfiles, nextLogs] = await Promise.all([
                GetConnectionStatus(),
                ListProfiles(),
                GetFrameLogs(),
            ]);
            setStatus(nextStatus);
            setProfileList(nextProfiles ?? []);
            setFrameLogs(nextLogs ?? []);
        } catch (err) {
            setError(errorMessage(err));
        }
    }

    async function refreshLogsAndStatus() {
        const [nextStatus, nextLogs] = await Promise.all([GetConnectionStatus(), GetFrameLogs()]);
        setStatus(nextStatus);
        setFrameLogs(nextLogs ?? []);
    }

    async function handleConnect() {
        await run('connect', async () => {
            const nextStatus = await Connect(connection);
            setStatus(nextStatus);
            setWritePreview(null);
            setNotice(`Connected to ${connection.host}:${connection.port}`);
        });
    }

    async function handleDisconnect() {
        await run('connect', async () => {
            const nextStatus = await Disconnect();
            setStatus(nextStatus);
            setWritePreview(null);
            setNotice('Disconnected');
        });
    }

    async function handleRead() {
        await run('read', async () => {
            if (isBitRead(readRequest.functionCode)) {
                const response = await ReadBits({
                    functionCode: readRequest.functionCode,
                    address: readRequest.address,
                    quantity: readRequest.quantity,
                    addressBase: readRequest.addressBase,
                });
                setBitRows(response.rows ?? []);
                setReadResultKind('bits');
                await refreshLogsAndStatus();
                setNotice(`Read ${response.rows?.length ?? 0} bit rows in ${response.durationMs.toFixed(1)} ms`);
                return;
            }

            const response = await ReadRegisters(readRequest);
            setRows(response.rows ?? []);
            setReadResultKind('registers');
            await refreshLogsAndStatus();
            setNotice(`Read ${response.rows?.length ?? 0} register rows in ${response.durationMs.toFixed(1)} ms`);
        });
    }

    async function handlePreviewWrite() {
        await run('write', async () => {
            const prepared = prepareWrite(writeRequest);
            const preview = await previewPreparedWrite(prepared);
            setWritePreview(preview);
            setNotice('Write preview refreshed');
        });
    }

    async function handleWrite() {
        await run('write', async () => {
            const prepared = prepareWrite(writeRequest);
            const preview = await previewPreparedWrite(prepared);
            setWritePreview(preview);
            const confirmed = window.confirm(
                [
                    writeSummary(prepared),
                    `Device: ${preview.host}:${preview.port}`,
                    `Unit ID: ${preview.unitId}`,
                    `Raw: ${preview.pduHex}`,
                    '',
                    'Continue?',
                ].join('\n'),
            );
            if (!confirmed) {
                setNotice('Write cancelled');
                return;
            }
            await executePreparedWrite(prepared);
            await refreshLogsAndStatus();
            setNotice(writeNotice(prepared));
        });
    }

    async function handleSaveProfile() {
        await run('profile', async () => {
            const profile: profiles.Profile = {
                version: 1,
                name: profileName.trim(),
                host: connection.host,
                port: connection.port,
                unitId: connection.unitId,
                timeoutMs: connection.timeoutMs,
                retries: connection.retries,
                addressBase: readRequest.addressBase,
                dataType: readRequest.dataType,
                byteOrder: readRequest.byteOrder,
            };
            await SaveProfile(profile);
            const nextProfiles = await ListProfiles();
            setProfileList(nextProfiles ?? []);
            setNotice(`Saved profile ${profile.name}`);
        });
    }

    async function handleDeleteProfile(name: string) {
        await run('profile', async () => {
            await DeleteProfile(name);
            const nextProfiles = await ListProfiles();
            setProfileList(nextProfiles ?? []);
            setNotice(`Deleted profile ${name}`);
        });
    }

    function handleLoadProfile(profile: profiles.Profile) {
        setConnection({
            host: profile.host,
            port: profile.port || 502,
            unitId: profile.unitId,
            timeoutMs: profile.timeoutMs || 3000,
            retries: profile.retries,
        });
        setReadRequest((current) => ({
            ...current,
            addressBase: profile.addressBase,
            dataType: profile.dataType || 'uint16',
            byteOrder: profile.byteOrder || 'ABCD',
        }));
        setProfileName(profile.name);
        setNotice(`Loaded profile ${profile.name}`);
    }

    async function run(scope: BusyState, action: () => Promise<void>) {
        setBusy(scope);
        setError('');
        try {
            await action();
        } catch (err) {
            const message = errorMessage(err);
            setError(message);
            setNotice('Action failed');
            await refreshLogsAndStatus().catch(() => undefined);
        } finally {
            setBusy('');
        }
    }

    return (
        <main className="app-shell">
            <aside className="sidebar">
                <div className="brand-block">
                    <div>
                        <p className="eyebrow">Modbus TCP Client</p>
                        <h1>ModbusDesk</h1>
                    </div>
                    <span className={status.connected ? 'status-pill connected' : 'status-pill'}>{status.connected ? 'Online' : 'Offline'}</span>
                </div>

                <ConnectionPanel
                    value={connection}
                    status={status}
                    busy={busy === 'connect'}
                    onChange={setConnection}
                    onConnect={handleConnect}
                    onDisconnect={handleDisconnect}
                />

                <ProfilesPanel
                    profiles={profileList}
                    profileName={profileName}
                    busy={busy === 'profile'}
                    connected={status.connected}
                    onProfileNameChange={setProfileName}
                    onSave={handleSaveProfile}
                    onLoad={handleLoadProfile}
                    onDelete={handleDeleteProfile}
                />
            </aside>

            <section className="workspace">
                <div className="topbar">
                    <div>
                        <span className="label">Endpoint</span>
                        <strong>{endpoint}</strong>
                    </div>
                    <div>
                        <span className="label">Unit</span>
                        <strong>{status.connected ? status.unitId : connection.unitId}</strong>
                    </div>
                    <div>
                        <span className="label">Last response</span>
                        <strong>{status.lastResponseMs ? `${status.lastResponseMs.toFixed(1)} ms` : '-'}</strong>
                    </div>
                    <div className={error ? 'message error' : 'message'}>{error || notice}</div>
                </div>

                <div className="action-grid">
                    <ReadPanel
                        value={readRequest}
                        busy={busy === 'read'}
                        connected={status.connected}
                        onChange={setReadRequest}
                        onRead={handleRead}
                    />
                    <WritePanel
                        value={writeRequest}
                        preview={writePreview}
                        busy={busy === 'write'}
                        connected={status.connected}
                        onChange={(next) => {
                            setWriteRequest(next);
                            setWritePreview(null);
                        }}
                        onPreview={handlePreviewWrite}
                        onWrite={handleWrite}
                    />
                </div>

                {readResultKind === 'bits' ? <BitTable rows={bitRows}/> : <RegisterTable rows={rows}/>}
                <FrameInspector logs={frameLogs}/>
            </section>
        </main>
    );
}

function isBitRead(functionCode: number) {
    return functionCode === 1 || functionCode === 2;
}

type PreparedWrite =
    | {functionCode: 5; request: modbus.WriteSingleCoilRequest}
    | {functionCode: 6; request: modbus.WriteSingleRegisterRequest}
    | {functionCode: 15; request: modbus.WriteMultipleCoilsRequest}
    | {functionCode: 16; request: modbus.WriteMultipleRegistersRequest};

function prepareWrite(form: WriteFormState): PreparedWrite {
    switch (form.functionCode) {
        case 5:
            return {
                functionCode: 5,
                request: {
                    address: form.address,
                    value: form.coilValue,
                },
            };
        case 6:
            return {
                functionCode: 6,
                request: {
                    address: form.address,
                    value: form.registerValue,
                    dataType: form.dataType,
                    byteOrder: form.byteOrder,
                },
            };
        case 15:
            return {
                functionCode: 15,
                request: {
                    address: form.address,
                    values: parseCoilValues(form.coilValues),
                },
            };
        case 16:
            return {
                functionCode: 16,
                request: {
                    address: form.address,
                    values: parseRegisterValues(form.registerValues),
                },
            };
        default:
            throw new Error(`Unsupported write function code ${form.functionCode}`);
    }
}

function previewPreparedWrite(prepared: PreparedWrite) {
    switch (prepared.functionCode) {
        case 5:
            return PreviewWriteSingleCoil(prepared.request);
        case 6:
            return PreviewWriteSingleRegister(prepared.request);
        case 15:
            return PreviewWriteMultipleCoils(prepared.request);
        case 16:
            return PreviewWriteMultipleRegisters(prepared.request);
    }
}

function executePreparedWrite(prepared: PreparedWrite) {
    switch (prepared.functionCode) {
        case 5:
            return WriteSingleCoil(prepared.request);
        case 6:
            return WriteSingleRegister(prepared.request);
        case 15:
            return WriteMultipleCoils(prepared.request);
        case 16:
            return WriteMultipleRegisters(prepared.request);
    }
}

function writeSummary(prepared: PreparedWrite) {
    switch (prepared.functionCode) {
        case 5:
            return `Write Coil ${prepared.request.address} = ${prepared.request.value ? 'ON' : 'OFF'}`;
        case 6:
            return `Write Holding Register ${prepared.request.address} = ${prepared.request.value}`;
        case 15:
            return `Write ${prepared.request.values.length} Coils from ${prepared.request.address}`;
        case 16:
            return `Write ${prepared.request.values.length} Holding Registers from ${prepared.request.address}`;
    }
}

function writeNotice(prepared: PreparedWrite) {
    switch (prepared.functionCode) {
        case 5:
            return `Wrote coil ${prepared.request.address}`;
        case 6:
            return `Wrote register ${prepared.request.address}`;
        case 15:
            return `Wrote ${prepared.request.values.length} coils`;
        case 16:
            return `Wrote ${prepared.request.values.length} registers`;
    }
}

function parseCoilValues(raw: string) {
    const tokens = splitValueTokens(raw);
    if (tokens.length === 0) {
        throw new Error('At least one coil value is required');
    }
    return tokens.map((token) => {
        const normalized = token.toLowerCase();
        if (['1', 'true', 'on', 'yes'].includes(normalized)) {
            return true;
        }
        if (['0', 'false', 'off', 'no'].includes(normalized)) {
            return false;
        }
        throw new Error(`Invalid coil value "${token}"`);
    });
}

function parseRegisterValues(raw: string) {
    const tokens = splitValueTokens(raw);
    if (tokens.length === 0) {
        throw new Error('At least one register value is required');
    }
    return tokens.map((token) => {
        const parsed = Number(token);
        if (!Number.isInteger(parsed) || parsed < 0 || parsed > 65535) {
            throw new Error(`Invalid register value "${token}"`);
        }
        return parsed;
    });
}

function splitValueTokens(raw: string) {
    return raw
        .split(/[\s,;]+/)
        .map((token) => token.trim())
        .filter(Boolean);
}

function errorMessage(err: unknown) {
    if (err instanceof Error) {
        return err.message;
    }
    if (typeof err === 'string') {
        return err;
    }
    return 'Unexpected error';
}

export default App;
