package liskPortal

const (
	TaskQuery = `
		mutation UpdateAirdropTaskStatus($input: UpdateTaskStatusInputData!) {
			userdrop {
				updateTaskStatus(input: $input) {
					success
					progress {
						isCompleted
						completedAt
					}
				}
			}
		}
	`
	TopQuery = `
		query AirdropUser($filter: UserFilter!, $pointsHistoryFilter: QueryFilter, $tasksFilter: QueryFilter) {
			userdrop {
				user(filter: $filter) {
					rank
					points
					updatedAt
					pointsHistories(filter: $pointsHistoryFilter) {
						totalCount
					}
					tasks(filter: $tasksFilter) {
						id
					}
				}
			}
		}
	`
)
