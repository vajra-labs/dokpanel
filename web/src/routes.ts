import {index, rootRoute, route} from '@tanstack/virtual-file-routes';

export const routes = rootRoute('_root.tsx', [
	index('index.tsx'),
	// Auth Pages
	route('singup', 'auth/singup.tsx'),
	route('singin', 'auth/singin.tsx'),
]);
