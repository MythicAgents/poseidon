from mythic_payloadtype_container.MythicCommandBase import *
import json
from mythic_payloadtype_container.MythicRPC import *


class ShellArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        pass


class ShellCommand(CommandBase):
    cmd = "shell"
    needs_admin = False
    help_cmd = "shell [command]"
    description = "Execute a shell command with 'bash -c'"
    version = 1
    author = "@xorrior"
    argument_class = ShellArguments
    attackmapping = ["T1059.004"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        resp = await MythicRPC().execute("create_artifact", task_id=task.id,
                                         artifact="/bin/bash -c {}".format(task.args.command_line),
                                         artifact_type="Process Create",
                                         )
        resp = await MythicRPC().execute("create_artifact", task_id=task.id,
                                         artifact="{}".format(task.args.command_line),
                                         artifact_type="Process Create",
                                         )
        return task

    async def process_response(self, response: AgentResponse):
        pass
