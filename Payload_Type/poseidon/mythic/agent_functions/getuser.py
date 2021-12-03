from mythic_payloadtype_container.MythicCommandBase import *
import json


class GetUserArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        pass


class GetUserCommand(CommandBase):
    cmd = "getuser"
    needs_admin = False
    help_cmd = "getuser"
    description = "Get information regarding the current user context."
    version = 1
    author = "@xorrior"
    argument_class = GetUserArguments
    attackmapping = ["T1033"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
