export function parseStringToJSON(raw: string) {
    const lines = raw.split('\n');
    const entries: Record<string, string> = {};
    return lines.reduce((acc: Record<string, string>, line) => {
        const [key, value] = line.split('\t');
        acc[key] = value;
        return acc;
    }, entries);
}
