import {modbus} from '../../wailsjs/go/models';
import {clampNumberInput} from '../lib/number';

type Props = {
    value: modbus.ReadRegistersRequest;
    busy: boolean;
    connected: boolean;
    onChange: (next: modbus.ReadRegistersRequest) => void;
    onRead: () => void;
};

const dataTypes = ['uint16', 'int16', 'float32', 'hex', 'binary'];
const byteOrders = ['ABCD', 'BADC', 'CDAB', 'DCBA'];

export function ReadPanel({value, busy, connected, onChange, onRead}: Props) {
    const bitRead = isBitRead(value.functionCode);
    const maxQuantity = bitRead ? 2000 : 125;

    return (
        <section className="panel action-panel">
            <div className="panel-header">
                <h2>Read</h2>
                <span className="function-code">FC{String(value.functionCode).padStart(2, '0')}</span>
            </div>
            <div className="field-row">
                <label>
                    Function
                    <select
                        value={value.functionCode}
                        onChange={(event) => {
                            const functionCode = Number(event.target.value);
                            const nextBitRead = isBitRead(functionCode);
                            onChange({
                                ...value,
                                functionCode,
                                quantity: nextBitRead ? value.quantity : Math.min(value.quantity, 125),
                            });
                        }}
                    >
                        <option value={1}>FC01 Coils</option>
                        <option value={2}>FC02 Discrete Inputs</option>
                        <option value={3}>FC03 Holding Registers</option>
                        <option value={4}>FC04 Input Registers</option>
                    </select>
                </label>
                <label>
                    Address base
                    <select
                        value={value.addressBase}
                        onChange={(event) => onChange({...value, addressBase: Number(event.target.value)})}
                    >
                        <option value={0}>0-based</option>
                        <option value={1}>1-based</option>
                    </select>
                </label>
            </div>
            <div className="field-row">
                <label>
                    Start address
                    <input
                        type="number"
                        min={0}
                        max={65535}
                        value={value.address}
                        onChange={(event) => onChange({...value, address: clampNumberInput(event.target.value, value.address, 0, 65535)})}
                    />
                </label>
                <label>
                    Quantity
                    <input
                        type="number"
                        min={1}
                        max={maxQuantity}
                        value={value.quantity}
                        onChange={(event) => onChange({...value, quantity: clampNumberInput(event.target.value, value.quantity, 1, maxQuantity)})}
                    />
                </label>
            </div>
            {!bitRead ? (
                <div className="field-row">
                    <label>
                        Data type
                        <select
                            value={value.dataType}
                            onChange={(event) => onChange({...value, dataType: event.target.value})}
                        >
                            {dataTypes.map((item) => <option value={item} key={item}>{item}</option>)}
                        </select>
                    </label>
                    <label>
                        Byte order
                        <select
                            value={value.byteOrder}
                            onChange={(event) => onChange({...value, byteOrder: event.target.value})}
                        >
                            {byteOrders.map((item) => <option value={item} key={item}>{item}</option>)}
                        </select>
                    </label>
                </div>
            ) : null}
            <button className="primary" disabled={busy || !connected} onClick={onRead}>Read once</button>
        </section>
    );
}

function isBitRead(functionCode: number) {
    return functionCode === 1 || functionCode === 2;
}
