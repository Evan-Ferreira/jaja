import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { formatDueDate } from '@/utils/string';
import { Course } from './types';
import { ResyncButton } from './resync-button';

async function getCourses(): Promise<Course[]> {
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/d2l/courses`);

    if (!res.ok) {
        const body = await res.text();
        console.error('[courses] error response body:', body);
        throw new Error(`Failed to load courses: ${res.status}`);
    }

    return await res.json();
}

export default async function CoursesPage() {
    let courses: Course[] = [];
    let error: string | null = null;

    try {
        courses = await getCourses();
    } catch (e) {
        error = e instanceof Error ? e.message : 'Unknown error';
    }

    return (
        <div className="min-h-screen p-8 max-w-4xl mx-auto">
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-bold">Courses & Assignments</h1>
                <ResyncButton />
            </div>

            {error && (
                <p className="text-sm text-red-500 border border-red-200 rounded px-4 py-3">
                    {error}
                </p>
            )}

            {!error && courses == null && (
                <p className="text-sm text-muted-foreground">No courses found.</p>
            )}

            <div className="flex flex-col gap-6">
                {courses.map((course) => (
                    <div key={course.id} className="border rounded-lg p-5">
                        <h2 className="font-semibold text-lg leading-tight">
                            {course.name}
                        </h2>
                        {course.code && (
                            <Label className="text-muted-foreground mt-1">
                                {course.code}
                            </Label>
                        )}

                        <Separator className="my-3" />

                        {course.assignments.length === 0 ? (
                            <p className="text-sm text-muted-foreground">
                                No assignments
                            </p>
                        ) : (
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="text-left">
                                        <th className="pb-2 font-medium text-muted-foreground">Assignment</th>
                                        <th className="pb-2 font-medium text-muted-foreground">Due</th>
                                        <th className="pb-2 font-medium text-muted-foreground text-right">Points</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {course.assignments.map((a) => (
                                        <tr key={a.id} className="border-t">
                                            <td className="py-2 pr-4">{a.name}</td>
                                            <td className="py-2 pr-4 whitespace-nowrap text-muted-foreground">
                                                {formatDueDate(a.due_date)}
                                            </td>
                                            <td className="py-2 text-right text-muted-foreground">
                                                {a.score_out_of ?? '—'}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
}
