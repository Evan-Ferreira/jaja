'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { syncCourses } from './actions';

export function ResyncButton() {
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    async function handleResync() {
        setLoading(true);
        setError(null);
        try {
            await syncCourses();
            router.refresh();
        } catch (e) {
            setError(e instanceof Error ? e.message : 'Unknown error');
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="flex flex-col items-end gap-1">
            <Button onClick={handleResync} disabled={loading}>
                {loading ? 'Syncing...' : 'Resync D2L'}
            </Button>
            {error && <p className="text-xs text-red-500">{error}</p>}
        </div>
    );
}
