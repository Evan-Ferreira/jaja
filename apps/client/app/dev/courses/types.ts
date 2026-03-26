export type Assignment = {
    id: number;
    name: string;
    instructions: string;
    due_date: string | null;
    score_out_of: number | null;
};

export type Course = {
    id: number;
    name: string;
    code: string;
    assignments: Assignment[];
};
