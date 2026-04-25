import {modbus} from '../../wailsjs/go/models';
import {clampNumberInput} from '../lib/number';

export type WriteFunctionCode = 5 | 6 | 15 | 16;

export type WriteFormState = {
    functionCode: WriteFunctionCode;
    address: number;
    coilValue: boolean;
    registerValue: number;
    coilValues: string;
    registerValues: string;
    dataType: string;
    byteOrder: string;
};

type Props = {
    value: WriteFormState;
    preview: modbus.WritePreview | null;
    busy: boolean;
    connected: boolean;
    onChange: (next: WriteFormState) => void;
    onPreview: () => void;
    onWrite: () => void;
};

export function WritePanel({value, preview, busy, connected, onChange, onPreview, onWrite}: Props) {
    const functionLabel = writeFunctionLabel(value.functionCode);

    return (
        <section className="panel action-panel write-panel">
            <div className="panel-header">
                <h2>Write</h2>
                <span className="function-code danger">FC{String(value.functionCode).padStart(2, '0')}</span>
            </div>
            <div className="field-row">
                <label>
                    Function
                    <select
                        value={value.functionCode}
                        onChange={(event) => onChange({
                            ...value,
                            functionCode: Number(event.target.value) as WriteFunctionCode,
                        })}
                    >
                        <option value={5}>FC05 Single Coil</option>
                        <option value={6}>FC06 Single Register</option>
                        <option value={15}>FC15 Multiple Coils</option>
                        <option value={16}>FC16 Multiple Registers</option>
                    </select>
                </label>
                <label>
                    Address
                    <input
                        type="number"
                        min={0}
                        max={65535}
                        value={value.address}
                        onChange={(event) => onChange({...value, address: clampNumberInput(event.target.value, value.address, 0, 65535)})}
                    />
                </label>
            </div>

            {value.functionCode === 5 ? (
                <label className="checkbox-row write-toggle">
                    <input
                        type="checkbox"
                        checked={value.coilValue}
                        onChange={(event) => onChange({...value, coilValue: event.target.checked})}
                    />
                    <span>Coil value</span>
                    <strong>{value.coilValue ? 'ON' : 'OFF'}</strong>
                </label>
            ) : null}

            {value.functionCode === 6 ? (
                <>
                    <div className="field-row">
                        <label>
                            UInt16 value
                            <input
                                type="number"
                                min={0}
                                max={65535}
                                value={value.registerValue}
                                onChange={(event) => onChange({
                                    ...value,
                                    registerValue: clampNumberInput(event.target.value, value.registerValue, 0, 65535),
                                })}
                            />
                        </label>
                    </div>
                    <div className="field-row">
                        <label>
                            Data type
                            <select
                                value={value.dataType || 'uint16'}
                                onChange={(event) => onChange({...value, dataType: event.target.value})}
                            >
                                <option value="uint16">uint16</option>
                                <option value="hex">hex</option>
                            </select>
                        </label>
                        <label>
                            Byte order
                            <select
                                value={value.byteOrder || 'ABCD'}
                                onChange={(event) => onChange({...value, byteOrder: event.target.value})}
                            >
                                <option value="ABCD">ABCD</option>
                            </select>
                        </label>
                    </div>
                </>
            ) : null}

            {value.functionCode === 15 ? (
                <div className="field-row stacked">
                    <label>
                        Coil values
                        <textarea
                            value={value.coilValues}
                            onChange={(event) => onChange({...value, coilValues: event.target.value})}
                            rows={3}
                        />
                    </label>
                </div>
            ) : null}

            {value.functionCode === 16 ? (
                <div className="field-row stacked">
                    <label>
                        Register values
                        <textarea
                            value={value.registerValues}
                            onChange={(event) => onChange({...value, registerValues: event.target.value})}
                            rows={3}
                        />
                    </label>
                </div>
            ) : null}

            <div className="preview-box">
                <span>Raw preview</span>
                <code>{preview?.pduHex || `${functionLabel} preview not generated`}</code>
                {preview?.frameHex ? <code className="muted-code">{preview.frameHex}</code> : null}
                {preview?.values?.length ? (
                    <div className="preview-values">
                        {preview.values.slice(0, 12).map((item, index) => (
                            <span key={`${item}-${index}`}>{item}</span>
                        ))}
                        {preview.values.length > 12 ? <span>+{preview.values.length - 12}</span> : null}
                    </div>
                ) : null}
            </div>
            <div className="button-row">
                <button disabled={busy || !connected} onClick={onPreview}>Preview</button>
                <button className="danger-button" disabled={busy || !connected} onClick={onWrite}>Write</button>
            </div>
        </section>
    );
}

function writeFunctionLabel(functionCode: WriteFunctionCode) {
    switch (functionCode) {
        case 5:
            return 'FC05';
        case 6:
            return 'FC06';
        case 15:
            return 'FC15';
        case 16:
            return 'FC16';
    }
}
