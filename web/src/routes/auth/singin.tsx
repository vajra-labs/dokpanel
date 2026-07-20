import {createFileRoute} from '@tanstack/react-router';

export const Route = createFileRoute('/singin')({
	component: RouteComponent,
});

function RouteComponent() {
	return <div>Hello "/singin"!</div>;
}
