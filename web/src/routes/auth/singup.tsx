import {createFileRoute} from '@tanstack/react-router';

export const Route = createFileRoute('/singup')({
	component: RouteComponent,
});

function RouteComponent() {
	return <div>Hello "/singup"!</div>;
}
