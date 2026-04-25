import {modbus} from '../../wailsjs/go/models';
import {clampNumberInput} from '../lib/number';

type Props = {
    value: modbus.ConnectRequest;
    status: modbus.ConnectionStatus;
    busy: boolean;
    onChange: (next: modbus.ConnectRequest) => void;
    onConnect: () => void;
    onDisconnect: () => void;
};

export function ConnectionPanel({value, status, busy, onChange, onConnect, onDisconnect}: Props) {
    return (
        <section className="panel">
            <div className="panel-header">
                <h2>Connection</h2>
                <span className={status.connected ? 'dot ok' : 'dot'}/>
            </div>
            <label>
                Host/IP
                <input
                    value={value.host}
                    disabled={status.connected}
                    onChange={(event) => onChange({...value, host: event.target.value})}
                    placeholder="192.168.2.30"
                />
            </label>
            <div className="field-row">
                <label>
                    Port
                    <input
                        type="number"
                        min={1}
                        max={65535}
                        value={value.port}
                        disabled={status.connected}
                        onChange={(event) => onChange({...value, port: clampNumberInput(event.target.value, value.port, 1, 65535)})}
                    />
                </label>
                <label>
                    Unit ID
                    <input
                        type="number"
                        min={0}
                        max={247}
                        value={value.unitId}
                        disabled={status.connected}
                        onChange={(event) => onChange({...value, unitId: clampNumberInput(event.target.value, value.unitId, 0, 247)})}
                    />
                </label>
            </div>
            <div className="field-row">
                <label>
                    Timeout ms
                    <input
                        type="number"
                        min={100}
                        value={value.timeoutMs}
                        disabled={status.connected}
                        onChange={(event) => onChange({...value, timeoutMs: clampNumberInput(event.target.value, value.timeoutMs, 100, 120000)})}
                    />
                </label>
                <label>
                    Retries
                    <input
                        type="number"
                        min={0}
                        max={10}
                        value={value.retries}
                        disabled={status.connected}
                        onChange={(event) => onChange({...value, retries: clampNumberInput(event.target.value, value.retries, 0, 10)})}
                    />
                </label>
            </div>
            <div className="button-row">
                <button className="primary" disabled={busy || status.connected} onClick={onConnect}>Connect</button>
                <button disabled={busy || !status.connected} onClick={onDisconnect}>Disconnect</button>
            </div>
        </section>
    );
}
