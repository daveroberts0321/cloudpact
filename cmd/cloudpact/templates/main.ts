interface User {
    id: string;
    name: string;
    email: string;
}

function init(): void {
    console.log('CloudPact frontend for {{.ProjectName}} loaded');
}

document.addEventListener('DOMContentLoaded', init);

