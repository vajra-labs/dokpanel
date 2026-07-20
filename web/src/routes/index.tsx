import {createFileRoute} from '@tanstack/react-router';
import React from 'react';

export const Route = createFileRoute('/')({component: Home});

function Home() {
	React.useEffect(() => {
		fetch('/api/health').then(async res => {
			const json = await res.json();
			console.log(json);
		});
	}, []);

	return (
		<div className="p-8">
			<h1 className="text-4xl font-bold">Welcome to TanStack Start</h1>
			<p className="mt-4 text-lg">
				Edit <code>src/routes/index.tsx</code> to get started.
			</p>
		</div>
	);
}
