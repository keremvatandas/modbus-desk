import {modbus} from '../../wailsjs/go/models';

type Props = {
    logs: modbus.FrameLog[];
};

export function FrameInspector({logs}: Props) {
    return (
        <section className="panel frame-panel">
            <div className="panel-header">
                <h2>Raw Frames</h2>
                <span className="count">{logs.length} frames</span>
            </div>
            <div className="frame-list">
                {logs.length === 0 ? (
                    <div className="empty-state">No TX/RX frames</div>
                ) : logs.slice().reverse().map((log) => (
                    <div className="frame-row" key={log.id}>
                        <span className={log.direction === 'TX' ? 'direction tx' : 'direction rx'}>{log.direction}</span>
                        <span>{log.timestamp}</span>
                        <span>TID {log.transactionId}</span>
                        <span>FC {formatCode(log.functionCode)}</span>
                        <code>{log.fullHex}</code>
                        <span>{log.durationMs ? `${log.durationMs.toFixed(1)} ms` : '-'}</span>
                        <button className="copy-frame" onClick={() => copyFrame(log.fullHex)}>Copy</button>
                        {log.exceptionCode ? <span className="exception">{log.exceptionCode}</span> : null}
                    </div>
                ))}
            </div>
        </section>
    );
}

function copyFrame(fullHex: string) {
    if (!navigator.clipboard) {
        return;
    }
    void navigator.clipboard.writeText(fullHex);
}

function formatCode(code: number) {
    return code.toString(16).toUpperCase().padStart(2, '0');
}
