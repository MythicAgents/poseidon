from mythic_payloadtype_container.MythicCommandBase import *
import json


class DrivesArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        pass


class DrivesCommand(CommandBase):
    cmd = "drives"
    needs_admin = False
    help_cmd = "drives"
    description = "Get information about mounted drives on Linux hosts only."
    version = 1
    author = "@xorrior"
    argument_class = DrivesArguments
    attributes = CommandAttributes(
        supported_os=[SupportedOS.Linux]
    )
    attackmapping = ["T1135"]

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
