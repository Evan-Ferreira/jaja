'use server';

export async function syncCourses() {
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/d2l/sync`, {
        method: 'POST',
    });
    if (!res.ok) throw new Error(`Sync failed: ${res.status}`);
}
