from mythic_payloadtype_container.MythicCommandBase import *
import json


class UnsetEnvArguments(TaskArguments):
    def __init__(self, command_line, **kwargs):
        super().__init__(command_line, **kwargs)
        self.args = []

    async def parse_arguments(self):
        if len(self.command_line) == 0:
            raise Exception("Must specify the environment variable to unset")


class UnsetEnvCommand(CommandBase):
    cmd = "unsetenv"
    needs_admin = False
    help_cmd = "unsetenv [param]"
    description = "Unset an environment variable"
    version = 1
    author = "@xorrior"
    argument_class = UnsetEnvArguments
    attackmapping = []
    attributes = CommandAttributes(
        supported_os=[SupportedOS.MacOS, SupportedOS.Linux]
    )

    async def create_tasking(self, task: MythicTask) -> MythicTask:
        return task

    async def process_response(self, response: AgentResponse):
        pass
