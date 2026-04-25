export function clampNumberInput(raw: string, fallback: number, min: number, max: number) {
    if (raw.trim() === '') {
        return fallback;
    }
    const parsed = Number(raw);
    if (!Number.isFinite(parsed)) {
        return fallback;
    }
    return Math.min(max, Math.max(min, Math.trunc(parsed)));
}
