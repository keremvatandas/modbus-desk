import {modbus} from '../../wailsjs/go/models';

type Props = {
    rows: modbus.RegisterRow[];
};

export function RegisterTable({rows}: Props) {
    return (
        <section className="panel table-panel">
            <div className="panel-header">
                <h2>Register Table</h2>
                <span className="count">{rows.length} rows</span>
            </div>
            <div className="table-wrap">
                <table>
                    <thead>
                    <tr>
                        <th>Address</th>
                        <th>Raw Hex</th>
                        <th>UInt16</th>
                        <th>Int16</th>
                        <th>Float32</th>
                        <th>Binary</th>
                        <th>Timestamp</th>
                        <th>Quality</th>
                    </tr>
                    </thead>
                    <tbody>
                    {rows.length === 0 ? (
                        <tr>
                            <td colSpan={8} className="empty-cell">No register data</td>
                        </tr>
                    ) : rows.map((row) => (
                        <tr key={`${row.address}-${row.timestamp}`}>
                            <td>{formatAddress(row)}</td>
                            <td><code>{row.rawHex}</code></td>
                            <td>{row.uint16}</td>
                            <td>{row.int16}</td>
                            <td>{formatFloat(row.float32)}</td>
                            <td><code>{row.binary}</code></td>
                            <td>{row.timestamp}</td>
                            <td><span className={row.quality === 'OK' ? 'quality ok' : 'quality warn'}>{row.quality}</span></td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>
        </section>
    );
}

function formatFloat(value?: number) {
    if (value === undefined || value === null) {
        return '-';
    }
    return Number.isInteger(value) ? value.toFixed(1) : value.toFixed(4);
}

function formatAddress(row: modbus.RegisterRow) {
    if (row.registerCount && row.registerCount > 1) {
        return `${row.displayAddress}-${row.displayAddress + row.registerCount - 1}`;
    }
    return String(row.displayAddress);
}
