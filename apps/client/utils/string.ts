export function parseStringToJSON(raw: string) {
    const lines = raw.split('\n');
    const entries: Record<string, string> = {};
    return lines.reduce((acc: Record<string, string>, line) => {
        const [key, value] = line.split('\t');
        acc[key] = value;
        return acc;
    }, entries);
}

export function formatDueDate(iso: string | null): string {
    if (!iso) return 'No due date';
    return new Date(iso).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
    });
}
