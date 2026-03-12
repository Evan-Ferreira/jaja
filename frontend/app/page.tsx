import CookieForm from './components/cookie-form';

export default function Home() {
    return (
        <div className="min-w-screen max-w-screen min-h-screen">
            <main className="flex w-full items-center justify-center min-h-screen">
                <CookieForm />
            </main>
        </div>
    );
}
