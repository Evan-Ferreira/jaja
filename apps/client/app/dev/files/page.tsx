'use client';

import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
    Field,
    FieldDescription,
    FieldGroup,
    FieldLabel,
    FieldLegend,
    FieldSet,
} from '@/components/ui/field';
import { toast } from 'sonner';
import { useState } from 'react';

export default function DevPage() {
    const [isLoading, setIsLoading] = useState(false);

    return (
        <div className="w-full h-full min-h-screen flex items-center justify-center">
            <form
                className="w-full max-w-md"
                onSubmit={async (e) => {
                    try {
                        e.preventDefault();
                        setIsLoading(true);
                        const formData = new FormData(e.currentTarget);

                        const res = await fetch(
                            `${process.env.NEXT_PUBLIC_API_URL}/dev/assignment-files`,
                            {
                                method: 'POST',
                                body: formData,
                            },
                        );

                        const data = await res.json();

                        if (!res.ok) {
                            throw new Error(
                                'Failed to upload files',
                                data.error,
                            );
                        }

                        toast.success('Files uploaded successfully');
                    } catch (error) {
                        console.error(error);
                        toast.error('Failed to upload files');
                    } finally {
                        setIsLoading(false);
                    }
                }}
            >
                <FieldGroup>
                    <FieldSet>
                        <FieldLegend>Assignment Files</FieldLegend>
                        <FieldDescription>
                            Upload the files for your assignment
                        </FieldDescription>
                        <FieldGroup>
                            <Field>
                                <FieldLabel htmlFor="assignment_instructions_rubric">
                                    Assignment Instructions/Rubric
                                </FieldLabel>
                                <Input
                                    placeholder="Assignment Instructions/Rubric"
                                    name="assignment_instructions_rubric"
                                    required
                                    type="file"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="assignment_instructions_rubric">
                                    File Name
                                </FieldLabel>
                                <Input
                                    placeholder="File Name"
                                    name="file_name"
                                    required
                                />
                            </Field>
                        </FieldGroup>
                    </FieldSet>
                    <Field orientation="vertical">
                        <Button type="submit" disabled={isLoading}>
                            {isLoading ? 'Uploading...' : 'Upload Files'}
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}
