import {modbus} from '../../wailsjs/go/models';

type Props = {
    rows: modbus.BitRow[];
};

export function BitTable({rows}: Props) {
    return (
        <section className="panel table-panel">
            <div className="panel-header">
                <h2>Bit Table</h2>
                <span className="count">{rows.length} rows</span>
            </div>
            <div className="table-wrap">
                <table>
                    <thead>
                    <tr>
                        <th>Address</th>
                        <th>Value</th>
                        <th>Raw</th>
                        <th>Timestamp</th>
                        <th>Quality</th>
                    </tr>
                    </thead>
                    <tbody>
                    {rows.length === 0 ? (
                        <tr>
                            <td colSpan={5} className="empty-cell">No bit data</td>
                        </tr>
                    ) : rows.map((row) => (
                        <tr key={`${row.address}-${row.timestamp}`}>
                            <td>{row.displayAddress}</td>
                            <td>
                                <span className={row.value ? 'bit-state on' : 'bit-state off'}>
                                    {row.value ? 'ON' : 'OFF'}
                                </span>
                            </td>
                            <td><code>{row.raw}</code></td>
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
