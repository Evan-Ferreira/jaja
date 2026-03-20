'use client';

import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
    Field,
    FieldDescription,
    FieldGroup,
    FieldLabel,
    FieldLegend,
    FieldSet,
} from '@/components/ui/field';
import { Textarea } from '@/components/ui/textarea';
import { parseStringToJSON } from '@/utils/string';
import { useState } from 'react';
import { toast } from 'sonner';

export default function CookieForm() {
    const [cookies, setCookies] = useState<string>('');
    const [localStorage, setLocalStorage] = useState<string>('');
    const [termsAccepted, setTermsAccepted] = useState<boolean>(false);

    return (
        <div className="w-full max-w-md">
            <form
                onSubmit={async (e) => {
                    try {
                        e.preventDefault();
                        if (!termsAccepted || !cookies || !localStorage) {
                            return;
                        }
                        const parsedCookies = parseStringToJSON(cookies);
                        const parsedLocalStorage =
                            parseStringToJSON(localStorage);

                        const response = await fetch(
                            `${process.env.NEXT_PUBLIC_API_URL}/api/d2l/auth`,
                            {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json',
                                },
                                body: JSON.stringify({
                                    cookies: parsedCookies,
                                    local_storage: parsedLocalStorage,
                                }),
                            },
                        );

                        if (!response.ok) {
                            throw new Error(
                                'Failed to send cookies and local storage',
                            );
                        }

                        setCookies('');
                        setLocalStorage('');
                        setTermsAccepted(false);

                        toast.success(
                            'Cookies and local storage sent successfully',
                        );
                    } catch (error) {
                        console.error(
                            'Failed to send cookies and local storage',
                            error,
                        );
                        toast.error('Failed to send cookies and local storage');
                    }
                }}
            >
                <FieldGroup>
                    <FieldSet>
                        <FieldLegend>D2L Cookies and Local Storage</FieldLegend>
                        <FieldDescription>
                            Paste your D2L authentication cookies and
                            credentials below.
                        </FieldDescription>
                    </FieldSet>
                    <FieldSet>
                        <FieldGroup>
                            <Field>
                                <FieldLabel htmlFor="cookies">
                                    Cookies
                                </FieldLabel>
                                <Textarea
                                    placeholder="Copy and paste your D2L cookies here"
                                    className="resize-none h-24"
                                    id="cookies"
                                    value={cookies}
                                    onChange={(e) => setCookies(e.target.value)}
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="local-storage">
                                    Local Storage
                                </FieldLabel>
                                <Textarea
                                    placeholder="Copy and paste your D2L local storage here"
                                    className="resize-none h-24"
                                    id="local-storage"
                                    value={localStorage}
                                    onChange={(e) =>
                                        setLocalStorage(e.target.value)
                                    }
                                />
                            </Field>
                        </FieldGroup>
                    </FieldSet>
                    <FieldSet>
                        <FieldGroup>
                            <Field orientation="horizontal">
                                <Checkbox
                                    id="terms-accepted"
                                    checked={termsAccepted}
                                    onCheckedChange={(checked) =>
                                        setTermsAccepted(checked === true)
                                    }
                                />
                                <FieldLabel htmlFor="terms-accepted">
                                    I have read and agree to the{' '}
                                    <a
                                        href="https://www.youtube.com/watch?v=dQw4w9WgXcQ"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="underline"
                                    >
                                        terms and conditions
                                    </a>
                                </FieldLabel>
                            </Field>
                        </FieldGroup>
                    </FieldSet>
                    <Field orientation="vertical">
                        <Button
                            type="submit"
                            disabled={
                                !termsAccepted || !cookies || !localStorage
                            }
                        >
                            Submit
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}
